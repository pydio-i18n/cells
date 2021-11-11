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

// Package grpc wraps a Minio server for exposing the content of the datasource with the S3 protocol.
package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/proto/object"
	"github.com/pydio/cells/v4/common/service"
)

func init() {

	plugins.Register("main", func(ctx context.Context) {

		//todo v4
		//sources := config.SourceNamesForDataServices(common.ServiceDataObjects)
		sources := []string{"local1"}

		for _, datasource := range sources {

			service.NewService(
				service.Name(common.ServiceGrpcNamespace_+common.ServiceDataObjects_+datasource),
				service.Context(ctx),
				service.Tag(common.ServiceTagDatasource),
				service.Description("S3 Object service for a given datasource"),
				service.Source(datasource),
				service.Fork(true),
				service.Unique(true),
				service.AutoStart(false),
				service.WithGRPC(func(server *grpc.Server) error {
					engine := &ObjectHandler{
						Config: &object.MinioConfig{
							Name:        datasource,
							ApiKey:      "mycustomapikey",
							ApiSecret:   "mycustomapisecret",
							StorageType: object.StorageType_LOCAL,
							LocalFolder: "/Users/charles/Library/Application Support/Pydio/cells/data",
						},
						MinioConsolePort: 8383,
					}
					object.RegisterObjectsEndpointServer(server, engine)
					go engine.StartMinioServer(ctx, datasource)
					return nil
				}),
				/*
					service.WithMicro(func(m micro.Service) error {
						s := m.Options().Server
						serviceName := s.Options().Metadata["source"]

						engine := &ObjectHandler{}

						object.RegisterObjectsEndpointHandler(s, engine)

						m.Init(
							micro.AfterStart(func() error {
								ctx := m.Options().Context
								log.Logger(ctx).Debug("AfterStart for Object service " + serviceName)
								var conf *object.MinioConfig
								if err := servicecontext.ScanConfig(ctx, &conf); err != nil {
									return err
								}

								if ip, e := net.GetExternalIP(); e != nil {
									conf.RunningHost = "127.0.0.1"
								} else {
									conf.RunningHost = ip.String()
								}

								conf.RunningSecure = false

								engine.Config = conf
								log.Logger(ctx).Debug("Now starting minio server (" + serviceName + ")")
								go engine.StartMinioServer(ctx, serviceName)

								return nil
							}),
						)

						return nil
					}),
				*/
			)
		}
	})
}
