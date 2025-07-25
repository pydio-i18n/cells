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

// Package service provides a GRPC persistence layer for user-defined metadata
package service

import (
	"context"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/broker"
	"github.com/pydio/cells/v5/common/errors"
	meta2 "github.com/pydio/cells/v5/common/nodes/meta"
	"github.com/pydio/cells/v5/common/proto/idm"
	service2 "github.com/pydio/cells/v5/common/proto/service"
	"github.com/pydio/cells/v5/common/proto/tree"
	"github.com/pydio/cells/v5/common/runtime"
	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/common/service"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/idm/meta"
	grpc2 "github.com/pydio/cells/v5/idm/meta/grpc"
)

var (
	Name = common.ServiceGrpcNamespace_ + common.ServiceUserMeta
)

func init() {
	runtime.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(Name),
			service.Context(ctx),
			service.Tag(common.ServiceTagIdm),
			service.Metadata(meta2.ServiceMetaProvider, "stream"),
			service.Metadata(meta2.ServiceMetaNsProvider, "list"),
			service.Metadata(meta2.ServiceMetaProviderRequired, "true"),
			service.Description("User-defined Metadata"),
			service.WithStorageDrivers(meta.Drivers),
			service.Migrations([]*service.Migration{
				{
					TargetVersion: service.FirstRun(),
					Up: func(ctx context.Context) error {
						dao, err := manager.Resolve[meta.DAO](ctx)
						if err != nil {
							return err
						}
						if err = dao.Migrate(ctx); err != nil {
							return err
						}
						return defaultMetas(ctx, dao)
					},
				},
			}),
			service.Unique(true),
			service.WithGRPC(func(ctx context.Context, server grpc.ServiceRegistrar) error {

				handler := grpc2.NewHandler(ctx)
				idm.RegisterUserMetaServiceServer(server, handler)
				tree.RegisterNodeProviderStreamerServer(server, handler)

				// Clean role on user deletion
				if e := broker.SubscribeCancellable(ctx, common.TopicIdmEvent, func(ctx context.Context, message broker.Message) error {
					ev := &idm.ChangeEvent{}
					if ctx, e := message.Unmarshal(ctx, ev); e == nil {
						return grpc2.HandleClean(ctx, ev, Name)
					}
					return nil
				}, broker.WithCounterName("idm_meta")); e != nil {
					return e
				}

				return nil
			}),
		)
	})
}

func defaultMetas(ctx context.Context, dao meta.DAO) error {
	err := dao.GetNamespaceDao().Add(ctx, &idm.UserMetaNamespace{
		Namespace:      common.MetaNamespaceUserspacePrefix + "tags",
		Label:          "Tags",
		Indexable:      true,
		JsonDefinition: "{\"type\":\"tags\"}",
		Policies: []*service2.ResourcePolicy{
			{Action: service2.ResourcePolicyAction_READ, Subject: "*", Effect: service2.ResourcePolicy_allow},
			{Action: service2.ResourcePolicyAction_WRITE, Subject: "*", Effect: service2.ResourcePolicy_allow},
		},
	})
	if err == nil {
		log.Logger(ctx).Info("Inserted default namespace for metadata")
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// This is a duplicate error, we ignore it
		return nil
	}
	return err
}
