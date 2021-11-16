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

// Package grpc provides a service for storing and CRUD-ing ACLs
package grpc

import (
	"context"

	"github.com/pydio/cells/v4/common/micro/broker"
	"github.com/pydio/cells/v4/common/nodes/meta"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/idm/acl"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGrpcNamespace_+common.ServiceAcl),
			service.Context(ctx),
			service.Tag(common.ServiceTagIdm),
			service.Description("Access Control List service"),
			service.WithStorage(acl.NewDAO, "idm_acl"),
			service.Migrations([]*service.Migration{
				{
					TargetVersion: service.ValidVersion("1.2.0"),
					Up:            UpgradeTo120,
				},
			}),
			service.Metadata(meta.ServiceMetaProvider, "stream"),
			service.WithGRPC(func(ctx context.Context, server *grpc.Server) error {

				handler := new(Handler)
				idm.RegisterACLServiceServer(server, handler)
				tree.RegisterNodeProviderStreamerServer(server, handler)

				// Clean acls on Ws or Roles deletion
				rCleaner := &WsRolesCleaner{Handler: handler}
				unsub1, e := broker.Subscribe(common.TopicIdmEvent, func(message broker.Message) error {
					ev := &idm.ChangeEvent{}
					if ct, e := message.Unmarshal(ev); e == nil {
						return rCleaner.Handle(ct, ev)
					}
					return nil
				})
				if e != nil {
					return e
				}

				nCleaner := newNodesCleaner(ctx, handler)
				unsub2, e := broker.Subscribe(common.TopicTreeChanges, func(message broker.Message) error {
					ev := &tree.NodeChangeEvent{}
					if ct, e := message.Unmarshal(ev); e == nil {
						return nCleaner.Handle(ct, ev)
					}
					return nil
				})

				go func() {
					<-ctx.Done()
					_ = unsub1()
					_ = unsub2()
				}()

				// For when it will be used: clean locks at startup
				//	dao := servicecontext.GetDAO(m.Options().Context).(acl.DAO)
				//	if dao != nil {
				//		q, _ := anypb.New(&idm.ACLSingleQuery{Actions: []*idm.ACLAction{{Name: permissions.AclLock.Name}}})
				//		if num, _ := dao.Del(&service2.Query{SubQueries: []*anypb.Any{q}}); num > 0 {
				//			log.Logger(m.Options().Context).Info(fmt.Sprintf("Cleaned %d locks in ACLs", num))
				//		}
				//	}

				return nil
			}),
		)
	})
}
