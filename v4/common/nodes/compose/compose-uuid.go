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

package compose

import (
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/acl"
	"github.com/pydio/cells/v4/common/nodes/core"
	"github.com/pydio/cells/v4/common/nodes/encryption"
	"github.com/pydio/cells/v4/common/nodes/events"
	"github.com/pydio/cells/v4/common/nodes/put"
	"github.com/pydio/cells/v4/common/nodes/uuid"
	"github.com/pydio/cells/v4/common/nodes/version"
)

func UuidClient(oo ...nodes.Option) nodes.Client {
	return NewClient(UuidComposer(oo...)...)
}

func UuidComposer(oo ...nodes.Option) []nodes.Option {
	return append(oo,
		nodes.WithCore(func(pool nodes.SourcesPool) nodes.Handler {
			exe := &core.Executor{}
			exe.SetClientsPool(pool)
			return exe
		}),
		acl.WithAccessList(),
		uuid.WithWorkspace(),
		uuid.WithDatasource(),
		events.WithAudit(),
		acl.WithFilter(),
		//events.WithRead(), why not?
		put.WithPutInterceptor(),
		acl.WithLock(),
		put.WithUploadLimiter(),
		acl.WithContentLockFilter(),
		acl.WithQuota(),

		version.WithVersions(),
		encryption.WithEncryption(),
		core.WithFlatInterceptor(),
	)
}

// newUuidRouter is the legacy constructor. Returns a new configured instance of a router
// that relies on nodes UUID rather than the usual Node path.
/*
func newUuidRouter(options nodes.RouterOptions) nodes.Client {
	handlers := []nodes.Handler{
		acl.NewAccessListHandler(options.AdminView),
		uuid.NewUuidNodeHandler(),
		uuid.NewUuidDataSourceHandler(),
	}

	if options.AuditEvent {
		handlers = append(handlers, &events.HandlerAuditEvent{})
	}

	if !options.AdminView {
		handlers = append(handlers, &acl.AclFilterHandler{})
	}
	handlers = append(handlers, &put.PutHandler{}) // adds a node precreation on PUT file request
	if !options.AdminView {
		handlers = append(handlers, &put.UploadLimitFilter{})
		handlers = append(handlers, &acl.AclLockFilter{})
		handlers = append(handlers, &acl.AclContentLockFilter{})
		handlers = append(handlers, &acl.AclQuotaFilter{})
	}
	handlers = append(handlers, &version.VersionHandler{})
	handlers = append(handlers, &encryption.EncryptionHandler{}) // retrieves encryption materials from encryption service
	handlers = append(handlers, &core.FlatStorageHandler{})
	handlers = append(handlers, &core.Executor{})

	pool := nodes.NewClientsPool(options.WatchRegistry)
	return nodes.NewRouter(pool, handlers)
}
*/
