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

// Package role is in charge of managing user roles
package role

import (
	"context"

	"github.com/pydio/cells/v5/common/proto/idm"
	service2 "github.com/pydio/cells/v5/common/proto/service"
	"github.com/pydio/cells/v5/common/service"
	"github.com/pydio/cells/v5/common/storage/sql/resources"
)

var Drivers = service.StorageDrivers{}

// DAO interface
type DAO interface {
	resources.DAO

	Migrate(ctx context.Context) error
	Add(ctx context.Context, role *idm.Role) (*idm.Role, bool, error)
	Delete(ctx context.Context, query service2.Enquirer) (numRows int64, e error)
	Search(ctx context.Context, query service2.Enquirer, output *[]*idm.Role) error
	Count(ctx context.Context, query service2.Enquirer) (int32, error)
}
