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

// Package dav provides a REST gateway to communicate with pydio backend via the webdav protocol.
package dav

import (
	"context"
	"net/http"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/compose"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

var (
	davRouter nodes.Client
)

// GetHandlers is public to let external package spinning a DAV http handler
func GetHandlers(ctx context.Context) (http.Handler, nodes.Client) {
	if davRouter == nil {
		davRouter = compose.PathClient(ctx, nodes.WithAuditEventsLogging(), nodes.WithSynchronousCaching(), nodes.WithSynchronousTasks())
	}
	handler := newHandler(ctx, davRouter)
	handler = servicecontext.HttpWrapperMeta(ctx, handler)
	return handler, davRouter
}

func init() {

	runtime.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGatewayDav),
			service.Context(ctx),
			service.Tag(common.ServiceTagGateway),
			service.Description("DAV Gateway to tree service"),
			service.WithHTTP(func(runtimeCtx context.Context, mux server.HttpMux) error {
				if davRouter == nil {
					davRouter = compose.PathClient(runtimeCtx, nodes.WithAuditEventsLogging(), nodes.WithSynchronousCaching(), nodes.WithSynchronousTasks())
				}
				handler := newHandler(runtimeCtx, davRouter)
				handler = servicecontext.HttpWrapperMeta(runtimeCtx, handler)
				mux.Handle("/dav/", handler)
				return nil
			}),
			service.WithHTTPStop(func(ctx context.Context, mux server.HttpMux) error {
				if m, ok := mux.(server.PatternsProvider); ok {
					m.DeregisterPattern("/dav/")
				}
				return nil
			}),
		)
	})
}
