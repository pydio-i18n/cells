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

// Package rest exposes a REST API to manage shared rooms.
package rest

import (
	"context"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service"
	"github.com/pydio/cells/v4/idm/share"
)

func init() {
	runtime.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceRestNamespace_+common.ServiceShare),
			service.Context(ctx),
			service.Tag(common.ServiceTagIdm),
			service.Description("REST access to shared rooms"),
			service.Migrations([]*service.Migration{
				{
					TargetVersion: service.ValidVersion("1.6.2"),
					Up: func(ctx context.Context) error {
						sc := share.NewClient(ctx, nil)
						return sc.ClearLostHiddenUsers(ctx)
					},
				},
			}),
			service.WithWeb(func(c context.Context) service.WebHandler {
				return NewSharesHandler(c)
			}),
		)
	})
}
