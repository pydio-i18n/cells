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

// Package api starts the actual WebSocket service
package api

import (
	"context"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/context/metadata"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/gateway/websocket"
)

var (
	ws   *websocket.WebsocketHandler
	chat *websocket.ChatHandler
	name = common.ServiceGatewayNamespace_ + common.ServiceWebSocket
)

func publicationContext(publication *broker.Message) context.Context {
	c := metadata.NewContext(context.Background(), publication.Header)
	c = servicecontext.WithServiceName(c, name)
	return c
}

func init() {

	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(name),
			service.Context(ctx),
			service.Tag(common.ServiceTagGateway),
			service.Fork(true),
			service.Dependency(common.ServiceGrpcNamespace_+common.ServiceChat, []string{}),
			service.Description("WebSocket server pushing event to the clients"),
			/*
				service.WithHTTP(func() http.Handler {

					// ctx := s.Options().Context

					ctx := context.TODO()

					ws = websocket.NewWebSocketHandler(ctx)
					chat = websocket.NewChatHandler(ctx)

					ws.EventRouter = compose.ReverseClient(nodes.WithRegistryWatch())

					gin.SetMode(gin.ReleaseMode)
					gin.DisableConsoleColor()
					Server := gin.New()
					Server.Use(gin.Recovery())
					Server.GET("/event", func(c *gin.Context) {
						ws.Websocket.HandleRequest(c.Writer, c.Request)
					})

					Server.GET("/chat", func(c *gin.Context) {
						chat.Websocket.HandleRequest(c.Writer, c.Request)
					})

					return Server
				}),
				service.AfterStart(func(_ service.Service) error {
					brok := defaults.Broker()

					brok.Subscribe(common.TopicTreeChanges, func(publication *broker.Message) error {
						var event tree.NodeChangeEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return ws.HandleNodeChangeEvent(publicationContext(publication), &event)
						}
						return nil
					})

					brok.Subscribe(common.TopicMetaChanges, func(publication *broker.Message) error {
						var event tree.NodeChangeEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return ws.HandleNodeChangeEvent(publicationContext(publication), &event)
						}
						return nil
					})

					brok.Subscribe(common.TopicJobTaskEvent, func(publication *broker.Message) error {
						var event jobs.TaskChangeEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return ws.BroadcastTaskChangeEvent(publicationContext(publication), &event)
						}
						return nil
					})

					brok.Subscribe(common.TopicIdmEvent, func(publication *broker.Message) error {
						var event idm.ChangeEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return ws.BroadcastIDMChangeEvent(publicationContext(publication), &event)
						}
						return nil
					})

					brok.Subscribe(common.TopicActivityEvent, func(publication *broker.Message) error {
						var event activity.PostActivityEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return ws.BroadcastActivityEvent(publicationContext(publication), &event)
						}
						return nil
					})

					brok.Subscribe(common.TopicChatEvent, func(publication *broker.Message) error {
						var event chat2.ChatEvent
						if e := proto.Unmarshal(publication.Body, &event); e == nil {
							return chat.BroadcastChatMessage(publicationContext(publication), &event)
						}
						return nil
					})

					return nil
				}),
			*/
		)

	})
}
