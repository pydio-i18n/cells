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

// Package grpc provides persistence layer for CRUD-ing roles
package grpc

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/storage"
	"github.com/pydio/cells/v4/common/utils/configx"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service"
	"github.com/pydio/cells/v4/common/sql/resources"
	"github.com/pydio/cells/v4/idm/role"
)

const ServiceName = common.ServiceGrpcNamespace_ + common.ServiceRole

func init() {
	runtime.Register("main", func(ctx context.Context) {
		var s service.Service

		s = service.NewService(
			service.Name(ServiceName),
			service.Context(ctx),
			service.Tag(common.ServiceTagIdm),
			service.Description("Roles Service"),
			service.TODOMigrations(func() []*service.Migration {
				return []*service.Migration{
					{
						TargetVersion: service.FirstRun(),
						Up:            InitRoles,
					}, {
						TargetVersion: service.ValidVersion("1.2.0"),
						Up:            UpgradeTo12,
					}, {
						TargetVersion: service.ValidVersion("4.1.99"),
						Up:            UpgradeTo4199,
					}, {
						TargetVersion: service.ValidVersion("4.2.1"),
						Up:            UpgradeTo421,
					},
				}
			}),
			//service.WithTODOStorage(role.NewDAO, commonsql.NewDAO,
			//	service.WithStoragePrefix("idm_role"),
			//	service.WithStorageSupport(mysql.Driver, sqlite.Driver),
			//),
		)

		var srv grpc.ServiceRegistrar
		if !server.Get(&srv) {
			panic("no grpc server available")
		}

				// Clean role on user deletion
				cleaner := NewCleaner(ctx, handler)
				if e := broker.SubscribeCancellable(ctx, common.TopicIdmEvent, func(message broker.Message) error {
					ic := &idm.ChangeEvent{}
					if ct, e := message.Unmarshal(ic); e == nil {
						return cleaner.Handle(ct, ic)
					}
					return nil
				}, broker.WithCounterName("role")); e != nil {
					return e
				}

		opts := configx.New()
		dao.Init(ctx, opts)

		handler := NewHandler(ctx, s, dao.(role.DAO))
		idm.RegisterRoleServiceServer(srv, handler)

		// Clean role on user deletion
		cleaner := NewCleaner(ctx, handler, service.DAOProvider[resources.DAO](s))
		if e := broker.SubscribeCancellable(ctx, common.TopicIdmEvent, func(ctx context.Context, message broker.Message) error {
			ic := &idm.ChangeEvent{}
			if e := message.Unmarshal(ic); e == nil {
				return cleaner.Handle(ctx, ic)
			}
			return nil
		}); e != nil {
			panic(e)
		}
	})
}
