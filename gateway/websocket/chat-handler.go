/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package websocket

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	lkauth "github.com/livekit/protocol/auth"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pydio/melody"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/auth"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/compose"
	nodescontext "github.com/pydio/cells/v4/common/nodes/context"
	"github.com/pydio/cells/v4/common/proto/chat"
	"github.com/pydio/cells/v4/common/proto/tree"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	json "github.com/pydio/cells/v4/common/utils/jsonx"
)

const (
	SessionRoomKey = "room"
)

type ChatHandler struct {
	ctx       context.Context
	Websocket *melody.Melody
	Pool      nodes.SourcesPool
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(ctx context.Context) *ChatHandler {
	w := &ChatHandler{ctx: ctx}
	w.Pool = nodescontext.GetSourcesPool(ctx)
	w.initHandlers(ctx)
	return w
}

// BroadcastChatMessage sends chat message to connected sessions
func (c *ChatHandler) BroadcastChatMessage(ctx context.Context, msg *chat.ChatEvent) error {

	//marshaller := &jsonpb.Marshaler{}
	var buff []byte
	var compareRoomId string

	if msg.Message != nil {

		if msg.Details == "DELETE" {
			wsMessage := &chat.WebSocketMessage{
				Type:    chat.WsMessageType_DELETE_MSG,
				Message: msg.Message,
			}
			buff, _ = protojson.Marshal(wsMessage)
		} else {
			buff, _ = protojson.Marshal(msg.Message)
		}

		compareRoomId = msg.Message.RoomUuid

	} else if msg.Room != nil {

		compareRoomId = msg.Room.Uuid
		wsMessage := &chat.WebSocketMessage{
			Type: chat.WsMessageType_ROOM_UPDATE,
			Room: msg.Room,
		}
		buff, _ = protojson.Marshal(wsMessage)

	} else {
		return fmt.Errorf("Event should provide at least a Msg or a Room")
	}

	return c.Websocket.BroadcastFilter(buff, func(session *melody.Session) bool {
		if session.IsClosed() {
			log.Logger(ctx).Error("Session is closed")
			return false
		}
		_, found := c.roomInSession(session, compareRoomId)
		return found
	})

}

func (c *ChatHandler) initHandlers(ctx context.Context) {

	c.Websocket = melody.New()
	c.Websocket.Config.MaxMessageSize = 2048

	c.Websocket.HandleError(func(session *melody.Session, i error) {
		if !strings.Contains(i.Error(), "close 1000 (normal)") {
			log.Logger(ctx).Debug("HandleError", zap.Error(i))
		}
		session.Set(SessionRoomKey, nil)
		ClearSession(session)
	})

	c.Websocket.HandleClose(func(session *melody.Session, i int, i2 string) error {
		session.Set(SessionRoomKey, nil)
		ClearSession(session)
		return nil
	})

	c.Websocket.HandleMessage(func(session *melody.Session, payload []byte) {

		msg := &Message{}
		e := json.Unmarshal(payload, msg)
		if e == nil {
			switch msg.Type {
			case MsgSubscribe:
				if msg.JWT == "" {
					session.CloseWithMsg(NewErrorMessageString("Empty JWT"))
					log.Logger(ctx).Debug("empty jwt")
					return
				}
				verifier := auth.DefaultJWTVerifier()
				_, claims, e := verifier.Verify(ctx, msg.JWT)
				if e != nil {
					log.Logger(ctx).Error("invalid jwt received from websocket connection")
					session.CloseWithMsg(NewErrorMessage(e))
					return
				}
				updateSessionFromClaims(ctx, session, claims, c.Pool)
				return

			case MsgUnsubscribe:
				ClearSession(session)
				return
			}
		}

		chatMsg := &chat.WebSocketMessage{}
		e = protojson.Unmarshal(payload, chatMsg)
		if e != nil {
			log.Logger(ctx).Debug("Could not unmarshal message", zap.Error(e))
			return
		}
		// SAVE CTX IN SESSION?
		// ctx := context.Background()
		log.Logger(ctx).Debug("Got Message", zap.Any("msg", chatMsg))
		var userName string
		if userData, ok := session.Get(SessionUsernameKey); !ok && userData != nil {
			log.Logger(ctx).Debug("Chat Message requires ws subscription first")
			return
		} else {
			userName, ok = userData.(string)
			if !ok {
				log.Logger(ctx).Debug("Chat Message requires ws subscription first")
				return
			}
		}

		chatClient := chat.NewChatServiceClient(grpc.GetClientConnFromCtx(c.ctx, common.ServiceChat))

		switch chatMsg.Type {

		case chat.WsMessageType_JOIN:

			sessRoom := &sessionRoom{}
			if readonly, e := c.auth(session, chatMsg.Room); e != nil {
				log.Logger(ctx).Error("Not authorized to join this room", zap.Error(e))
				break
			} else {
				sessRoom.readonly = readonly
			}
			var isPing bool
			if chatMsg.Message != nil && chatMsg.Message.Message == "PING" {
				isPing = true
			}
			foundRoom, e1 := c.findOrCreateRoom(ctx, chatMsg.Room, !isPing)
			if e1 != nil || foundRoom == nil {
				log.Logger(ctx).Debug("CANNOT JOIN", zap.Any("msg", chatMsg), zap.Any("r", foundRoom), zap.Error(e1))
				break
			}
			sessRoom.uuid = foundRoom.Uuid
			c.heartbeat(ctx, userName, foundRoom)
			c.storeSessionRoom(session, sessRoom)
			// Update Room Users
			if save := c.appendUserToRoom(foundRoom, userName); save {

				_, e := chatClient.PutRoom(ctx, &chat.PutRoomRequest{Room: foundRoom})
				if e != nil {
					log.Logger(ctx).Error("Error while putting room", zap.Error(e))
				}
			}

		case chat.WsMessageType_LEAVE:

			foundRoom, e1 := c.findOrCreateRoom(ctx, chatMsg.Room, false)
			if e1 == nil && foundRoom != nil {
				if save := c.removeUserFromRoom(foundRoom, userName); save {
					chatClient.PutRoom(ctx, &chat.PutRoomRequest{Room: foundRoom})
					log.Logger(ctx).Debug("LEAVE", zap.Any("msg", chatMsg), zap.Any("r", foundRoom))
				}
				c.removeSessionRoom(session, foundRoom.Uuid)
			}

		case chat.WsMessageType_HISTORY:
			// Must arrive AFTER a JOIN message
			foundRoom, e1 := c.findOrCreateRoom(ctx, chatMsg.Room, false)
			if e1 != nil || foundRoom == nil {
				break
			}
			c.sendVideoInfoIfSupported(ctx, foundRoom.Uuid, session)
			request := &chat.ListMessagesRequest{RoomUuid: foundRoom.Uuid}
			if chatMsg.Message != nil {
				var offData map[string]int
				offsetMsg := chatMsg.Message.Message
				if e := json.Unmarshal([]byte(offsetMsg), &offData); e == nil {
					if offset, ok := offData["Offset"]; ok {
						request.Offset = int64(offset)
					}
					if limit, ok := offData["Limit"]; ok {
						request.Limit = int64(limit)
					}
				}
			}
			// List existing Messages
			ct, ca := context.WithCancel(ctx)
			defer ca()
			stream, e2 := chatClient.ListMessages(ct, request)
			if e2 == nil {
				for {
					resp, e3 := stream.Recv()
					if e3 != nil {
						break
					}
					b, _ := protojson.Marshal(resp.Message)
					session.Write(b)
				}
			}

		case chat.WsMessageType_POST:

			log.Logger(ctx).Debug("POST", zap.Any("msg", chatMsg))
			if session, found := c.roomInSession(session, chatMsg.Message.RoomUuid); !found || session.readonly {
				log.Logger(ctx).Error("Not authorized to post in this room")
				break
			}
			message := chatMsg.Message
			message.Author = userName
			message.Message = bluemonday.UGCPolicy().Sanitize(message.Message)
			message.Timestamp = time.Now().Unix()
			_, e := chatClient.PostMessage(ctx, &chat.PostMessageRequest{
				Messages: []*chat.ChatMessage{message},
			})
			if e != nil {
				log.Logger(ctx).Error("Error while posting message", zap.Any("msg", message), zap.Error(e))
			}

		case chat.WsMessageType_DELETE_MSG:

			log.Logger(ctx).Debug("Delete", zap.Any("msg", chatMsg))
			if session, found := c.roomInSession(session, chatMsg.Message.RoomUuid); !found || session.readonly {
				log.Logger(ctx).Error("Not authorized to post in this room")
				break
			}
			message := chatMsg.Message
			if message.Author == userName {
				_, e := chatClient.DeleteMessage(ctx, &chat.DeleteMessageRequest{
					Messages: []*chat.ChatMessage{message},
				})
				if e != nil {
					log.Logger(ctx).Error("Error while deleting message", zap.Any("msg", message), zap.Error(e))
				}
			}

		}

	})

}

type sessionRoom struct {
	uuid     string
	readonly bool
}

func (c *ChatHandler) roomInSession(session *melody.Session, roomUuid string) (*sessionRoom, bool) {
	if key, ok := session.Get(SessionRoomKey); ok && key != nil {
		rooms := key.([]*sessionRoom)
		for _, v := range rooms {
			if v.uuid == roomUuid {
				return v, true
			}
		}
		log.Logger(context.Background()).Debug("looking for rooms in session", zap.String("search", roomUuid))
	}
	return nil, false
}

func (c *ChatHandler) storeSessionRoom(session *melody.Session, room *sessionRoom) *melody.Session {
	var rooms []*sessionRoom
	if key, ok := session.Get(SessionRoomKey); ok && key != nil {
		rooms = key.([]*sessionRoom)
	}
	found := false
	for _, v := range rooms {
		if v.uuid == room.uuid {
			found = true
		}
	}
	if !found {
		rooms = append(rooms, room)
		log.Logger(context.Background()).Debug("storing rooms to session", zap.Any("room", room.uuid), zap.Int("rooms length", len(rooms)))
		session.Set(SessionRoomKey, rooms)
	} else {
		log.Logger(context.Background()).Debug("rooms to session already found", zap.Any("room", room.uuid), zap.Int("rooms length", len(rooms)))
	}
	return session
}

func (c *ChatHandler) removeSessionRoom(session *melody.Session, roomUuid string) *melody.Session {
	var rooms []*sessionRoom
	if key, ok := session.Get(SessionRoomKey); ok && key != nil {
		rooms = key.([]*sessionRoom)
	}
	var newRooms []*sessionRoom
	for _, k := range rooms {
		if k.uuid != roomUuid {
			newRooms = append(newRooms, k)
		}
	}
	log.Logger(context.Background()).Debug("removing room from session", zap.Any("room", roomUuid), zap.Int("rooms length", len(newRooms)))
	session.Set(SessionRoomKey, newRooms)
	return session
}

func (c *ChatHandler) findOrCreateRoom(ctx context.Context, room *chat.ChatRoom, createIfNotExists bool) (*chat.ChatRoom, error) {

	chatClient := chat.NewChatServiceClient(grpc.GetClientConnFromCtx(ctx, common.ServiceChat))
	ct, ca := context.WithCancel(ctx)
	defer ca()
	s, e := chatClient.ListRooms(ct, &chat.ListRoomsRequest{
		ByType:     room.Type,
		TypeObject: room.RoomTypeObject,
	})
	if e != nil {
		return nil, e
	}
	for {
		resp, rE := s.Recv()
		if rE != nil {
			break
		}
		if resp == nil {
			continue
		}
		return resp.Room, nil
	}

	if !createIfNotExists {
		return nil, nil
	}

	// if not returned yet, create
	resp, e1 := chatClient.PutRoom(ctx, &chat.PutRoomRequest{Room: room})
	if e1 != nil {
		return nil, e1
	}
	if resp.Room == nil {
		return nil, fmt.Errorf("nil room in response, this is not normal")
	}
	return resp.Room, nil

}

func (c *ChatHandler) appendUserToRoom(room *chat.ChatRoom, userName string) bool {
	uniq := map[string]string{}
	for _, u := range room.Users {
		uniq[u] = u
	}
	if _, already := uniq[userName]; already {
		return false
	}
	uniq[userName] = userName
	room.Users = []string{}
	for _, name := range uniq {
		room.Users = append(room.Users, name)
	}
	return true
}

func (c *ChatHandler) removeUserFromRoom(room *chat.ChatRoom, userName string) bool {
	users := []string{}
	var found bool
	for _, u := range room.Users {
		if u == userName {
			found = true
		} else {
			users = append(users, u)
		}
	}
	room.Users = users
	return found
}

var uuidRouter nodes.Client

// auth check authorization for the room. perm can be "join" or "post"
func (c *ChatHandler) auth(session *melody.Session, room *chat.ChatRoom) (bool, error) {

	var readonly bool
	ctx, err := prepareRemoteContext(c.ctx, session)
	if err != nil {
		return false, err
	}

	switch room.Type {
	case chat.RoomType_NODE:

		// Check node is readable and writeable
		if uuidRouter == nil {
			uuidRouter = compose.UuidClient(c.ctx)
		}
		resp, e := uuidRouter.ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{Uuid: room.RoomTypeObject}})
		if e != nil {
			return false, e
		}
		if _, er := uuidRouter.CanApply(ctx, &tree.NodeChangeEvent{Type: tree.NodeChangeEvent_UPDATE_CONTENT, Target: resp.Node}); er != nil {
			readonly = true
		}

	case chat.RoomType_USER:
		// Check that this user is visible to current user
	case chat.RoomType_WORKSPACE:
		// Check that workspace is accessible
	case chat.RoomType_GLOBAL:
		// TODO
	}
	return readonly, nil
}

func (c *ChatHandler) sendVideoInfoIfSupported(ctx context.Context, roomUuid string, session *melody.Session) {
	if os.Getenv("CELLS_ENABLE_LIVEKIT") == "" {
		return
	}
	conf := config.Get("frontend", "plugin", "action.livekit")
	if !conf.Val(config.KeyFrontPluginEnabled).Bool() {
		return
	}
	var lkUrl string
	if mc, ok := session.Get(SessionMetaContext); ok {
		meta := mc.(metadata.Metadata)
		if host, o := meta[servicecontext.HttpMetaHost]; o && host != "" {
			lkUrl = "wss://" + host
		}
	}
	if lkUrl == "" {
		return
	}
	apiKey := conf.Val("LK_API_KEY").String()
	apiSecret := conf.Val("LK_API_SECRET").String()
	apiSecret = config.Vault().Val(apiSecret).String()
	sessionUser, _ := session.Get(SessionUsernameKey)

	if token, e := c.getLKJoinToken(apiKey, apiSecret, roomUuid, sessionUser.(string)); e == nil {
		type CallData struct {
			Type     string `json:"@type"`
			RoomUuid string `json:"RoomUuid"`
			Url      string `json:"Url"`
			Token    string `json:"Token"`
		}
		cd, _ := json.Marshal(&CallData{Type: "VIDEO_CALL", RoomUuid: roomUuid, Url: lkUrl, Token: token})
		session.Write(cd)
	} else {
		log.Logger(ctx).Error("Cannot load LK Token")
	}

}

// getLKJoinToken computes a valid token for Livekit server
func (c *ChatHandler) getLKJoinToken(apiKey, apiSecret, room, identity string) (string, error) {
	at := lkauth.NewAccessToken(apiKey, apiSecret)
	p := true
	grant := &lkauth.VideoGrant{
		RoomCreate:     true,
		CanPublish:     &p,
		CanPublishData: &p,
		CanSubscribe:   &p,
		RoomJoin:       true,
		Room:           room,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}
