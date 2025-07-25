/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
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

package grpc

import (
	"io"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/pydio/cells/v5/broker/activity"
	"github.com/pydio/cells/v5/common/client/commons/treec"
	"github.com/pydio/cells/v5/common/errors"
	proto "github.com/pydio/cells/v5/common/proto/activity"
	"github.com/pydio/cells/v5/common/proto/tree"
	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/common/telemetry/log"
)

type Handler struct {
	proto.UnimplementedActivityServiceServer
}

func (h *Handler) PostActivity(stream proto.ActivityService_PostActivityServer) error {
	ctx := stream.Context()

	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return err
	}
	for {
		request, e := stream.Recv()
		if e == io.EOF {
			return nil
		}
		if e != nil && e != io.EOF {
			return e
		}
		var boxName activity.BoxName
		switch request.BoxName {
		case "inbox":
			boxName = activity.BoxInbox
		case "outbox":
			boxName = activity.BoxOutbox
		default:
			return errors.New("unrecognized box name")
		}
		if e := dao.PostActivity(ctx, request.OwnerType, request.OwnerId, boxName, request.Activity, true); e != nil {
			return e
		}
	}
}

func (h *Handler) StreamActivities(request *proto.StreamActivitiesRequest, stream proto.ActivityService_StreamActivitiesServer) error {

	ctx := stream.Context()
	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return err
	}
	log.Logger(ctx).Debug("Should get activities", zap.Any("r", request))
	treeStreamer := treec.NodeProviderStreamerClient(ctx) // tree.NewNodeProviderStreamerClient(grpc.ResolveConn(ctx, common.ServiceTree))
	sClient, e := treeStreamer.ReadNodeStream(ctx)
	if e != nil {
		return e
	}
	replace := make(map[string]string)
	valid := make(map[string]bool)

	result := make(chan *proto.Object)
	done := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case ac := <-result:
				if ac.Type != proto.ObjectType_Delete && ac.Object != nil && (ac.Object.Type == proto.ObjectType_Document || ac.Object.Type == proto.ObjectType_Folder) && ac.Object.Id != "" {
					oName := ac.Object.Name
					if _, o := valid[oName]; o {
						//fmt.Println("nothing to do ")
					} else if r, o := replace[oName]; o {
						//fmt.Println("replace from cache")
						ac.Object.Name = r
					} else if e := sClient.Send(&tree.ReadNodeRequest{Node: &tree.Node{Uuid: ac.Object.Id}}); e == nil {
						rsp, er := sClient.Recv()
						if er == nil {
							nP := strings.TrimRight(rsp.GetNode().GetPath(), "/")
							if oName != nP {
								//fmt.Println("replacing", oName, "with", nP)
								ac.Object.Name = nP
								replace[oName] = rsp.GetNode().GetPath()
							} else {
								//fmt.Println("set valid")
								valid[oName] = true
							}
						}
					}
				}
				stream.Send(&proto.StreamActivitiesResponse{
					Activity: ac,
				})
			case <-done:
				return
			}
		}
	}()

	boxName := activity.BoxOutbox
	if request.BoxName == "inbox" {
		boxName = activity.BoxInbox
	}

	var er error

	if request.Context == proto.StreamContext_NODE_ID {
		er = dao.ActivitiesFor(ctx, proto.OwnerType_NODE, request.ContextData, boxName, "", request.Offset, request.Limit, "", result, done)
		wg.Wait()
	} else if request.Context == proto.StreamContext_USER_ID {
		var refBoxOffset activity.BoxName
		if request.AsDigest {
			refBoxOffset = activity.BoxLastSent
		}
		er = dao.ActivitiesFor(ctx, proto.OwnerType_USER, request.ContextData, boxName, refBoxOffset, request.Offset, request.Limit, "", result, done)
		wg.Wait()
	}

	return er
}

func (h *Handler) Subscribe(ctx context.Context, request *proto.SubscribeRequest) (*proto.SubscribeResponse, error) {

	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return nil, err
	}
	if e := dao.UpdateSubscription(ctx, request.Subscription); e != nil {
		return nil, e
	}
	return &proto.SubscribeResponse{
		Subscription: request.Subscription,
	}, nil

}

func (h *Handler) SearchSubscriptions(request *proto.SearchSubscriptionsRequest, stream proto.ActivityService_SearchSubscriptionsServer) error {

	ctx := stream.Context()
	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return err
	}
	var userId string
	var objectType = proto.OwnerType_NODE
	if len(request.ObjectIds) == 0 {
		return errors.New("please provide one or more object id")
	}
	if len(request.UserIds) > 0 {
		userId = request.UserIds[0]
	}
	users, err := dao.ListSubscriptions(ctx, objectType, request.ObjectIds)
	if err != nil {
		return err
	}
	for _, sub := range users {
		if len(sub.Events) == 0 {
			continue
		}
		if userId != "" && sub.UserId != userId {
			continue
		}
		stream.Send(&proto.SearchSubscriptionsResponse{
			Subscription: sub,
		})
	}
	return nil
}

func (h *Handler) UnreadActivitiesNumber(ctx context.Context, request *proto.UnreadActivitiesRequest) (*proto.UnreadActivitiesResponse, error) {

	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return nil, err
	}
	number := dao.CountUnreadForUser(ctx, request.UserId)
	return &proto.UnreadActivitiesResponse{
		Number: int32(number),
	}, nil

}

func (h *Handler) SetUserLastActivity(ctx context.Context, request *proto.UserLastActivityRequest) (*proto.UserLastActivityResponse, error) {

	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return nil, err
	}
	var boxName activity.BoxName
	if request.BoxName == "lastread" {
		boxName = activity.BoxLastRead
	} else if request.BoxName == "lastsent" {
		boxName = activity.BoxLastSent
	} else {
		return nil, errors.New("invalid box name")
	}

	if err := dao.StoreLastUserInbox(ctx, request.UserId, boxName, request.ActivityId); err == nil {
		return &proto.UserLastActivityResponse{Success: true}, nil
	} else {
		return nil, err
	}

}

func (h *Handler) PurgeActivities(ctx context.Context, request *proto.PurgeActivitiesRequest) (*proto.PurgeActivitiesResponse, error) {

	dao, err := manager.Resolve[activity.DAO](ctx)
	if err != nil {
		return nil, err
	}
	if request.BoxName != string(activity.BoxInbox) && request.BoxName != string(activity.BoxOutbox) {
		return nil, errors.WithMessage(errors.InvalidParameters, "Please provide one of inbox|outbox box name")
	}
	count := int32(0)
	logger := func(s string, i int) {
		count += int32(i)
		log.TasksLogger(ctx).Info(s)
	}

	var updated time.Time
	if request.UpdatedBeforeTimestamp > 0 {
		updated = time.Unix(int64(request.UpdatedBeforeTimestamp), 0)
	}

	e := dao.Purge(ctx, logger, request.OwnerType, request.OwnerID, activity.BoxName(request.BoxName), int(request.MinCount), int(request.MaxCount), updated, request.CompactDB, request.ClearBackups)
	return &proto.PurgeActivitiesResponse{
		Success:      true,
		DeletedCount: count,
	}, e

}
