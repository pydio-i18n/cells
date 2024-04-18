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

// Package wopi serves files using the WOPI protocol.
package wopi

import (
	"context"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/compose"
	"github.com/pydio/cells/v4/common/runtime"
	routes "github.com/pydio/cells/v4/common/server/http/routes"
	"github.com/pydio/cells/v4/common/service"
)

var (
	client nodes.Client
)

const (
	RouteWOPI = "wopi"
)

func init() {
	routes.DeclareRoute(RouteWOPI, "WOPI API service", "/wopi")

	runtime.Register("main", func(ctx context.Context) {
		//RegisterMainAPIEndpoint(ctx ... "a" ... "/api/")
		service.NewService(
			service.Name(common.ServiceGatewayWopi),
			service.Context(ctx),
			service.Tag(common.ServiceTagGateway),
			service.Description("WOPI REST Gateway to tree service"),
			//service.RouterDependencies(),
			service.WithHTTP(func(ctx context.Context, mux routes.RouteRegistrar) error {
				client = compose.UuidClient(ctx, nodes.WithAuditEventsLogging())
				wopiRouter := NewRouter()
				mux.Route(RouteWOPI).Handle("/", wopiRouter, routes.WithStripPrefix())
				return nil
			}),
			service.WithHTTPStop(func(ctx context.Context, mux routes.RouteRegistrar) error {
				mux.DeregisterRoute(RouteWOPI)
				return nil
			}),
		)
	})
}
