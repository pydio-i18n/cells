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
	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/nodes"
	"github.com/pydio/cells/common/nodes/acl"
	"github.com/pydio/cells/common/nodes/archive"
	"github.com/pydio/cells/common/nodes/binaries"
	"github.com/pydio/cells/common/nodes/core"
	"github.com/pydio/cells/common/nodes/encryption"
	"github.com/pydio/cells/common/nodes/events"
	"github.com/pydio/cells/common/nodes/path"
	"github.com/pydio/cells/common/nodes/put"
	"github.com/pydio/cells/common/nodes/sync"
	"github.com/pydio/cells/common/nodes/version"
	"github.com/pydio/cells/common/nodes/virtual"
)

func init() {
	// abstract.AdminClient = NewStandardRouter(nodes.RouterOptions{AdminView: true, WatchRegistry: true})
}

func PathComposer(oo ...nodes.Option) []nodes.Option {
	return append(oo,
		nodes.WithCore(func(pool nodes.SourcesPool) nodes.Client {
			exe := &core.Executor{}
			exe.SetClientsPool(pool)
			return exe
		}),
		acl.WithAccessList(),
		binaries.WithBinaryStore(common.PydioThumbstoreNamespace, true, false, false),
		binaries.WithBinaryStore(common.PydioDocstoreBinariesNamespace, false, true, true),
		archive.WithArchives(),
		path.WithWorkspace(),
		path.WithMultipleRoots(),
		virtual.WithResolver(), // !options.BrowseVirtualNodes && !options.AdminView
		virtual.WithBrowser(),  // options.BrowseVirtualNodes
		path.WithRootResolver(),
		path.WithDatasource(),
		sync.WithCache(), // options.SynchronousCache
		events.WithAudit(),
		acl.WithFilter(),
		events.WithRead(),
		put.WithPutInterceptor(),
		acl.WithLock(),
		put.WithUploadLimiter(),
		acl.WithContentLockFilter(),
		acl.WithQuota(),
		sync.WithFolderTasks(), // options.SynchronousTasks
		version.WithVersions(),
		encryption.WithEncryption(),
		core.WithFlatInterceptor(),
	)
}

// NewStandardRouter returns a new configured instance of the default standard router.
func NewStandardRouter(options nodes.RouterOptions) *nodes.Router {

	handlers := []nodes.Client{
		acl.NewAccessListHandler(options.AdminView),
		&binaries.BinaryStoreHandler{
			StoreName:      common.PydioThumbstoreNamespace, // Direct access to dedicated Bucket for thumbnails
			TransparentGet: true,
		},
		&binaries.BinaryStoreHandler{
			StoreName:     common.PydioDocstoreBinariesNamespace, // Direct access to dedicated Bucket for pydio binaries
			AllowPut:      true,
			AllowAnonRead: true,
		},
	}
	handlers = append(handlers, archive.NewArchiveHandler())
	handlers = append(handlers, path.NewPathWorkspaceHandler())
	handlers = append(handlers, path.NewPathMultipleRootsHandler())
	if !options.BrowseVirtualNodes && !options.AdminView {
		handlers = append(handlers, virtual.NewVirtualNodesHandler())
	}
	if options.BrowseVirtualNodes {
		handlers = append(handlers, virtual.NewVirtualNodesBrowser())
	}
	handlers = append(handlers, path.NewWorkspaceRootResolver())
	handlers = append(handlers, path.NewPathDataSourceHandler())

	if options.SynchronousCache {
		handlers = append(handlers, sync.NewSynchronousCacheHandler())
	}
	if options.AuditEvent {
		handlers = append(handlers, &events.HandlerAuditEvent{})
	}
	if !options.AdminView {
		handlers = append(handlers, &acl.AclFilterHandler{})
	}
	if options.LogReadEvents {
		handlers = append(handlers, &events.HandlerEventRead{})
	}

	handlers = append(handlers, &put.PutHandler{})
	handlers = append(handlers, &acl.AclLockFilter{})
	if !options.AdminView {
		handlers = append(handlers, &put.UploadLimitFilter{})
		handlers = append(handlers, &acl.AclContentLockFilter{})
		handlers = append(handlers, &acl.AclQuotaFilter{})
	}

	if options.SynchronousTasks {
		handlers = append(handlers, &sync.SyncFolderTasksHandler{})
	}
	handlers = append(handlers, &version.VersionHandler{})
	handlers = append(handlers, &encryption.EncryptionHandler{})
	handlers = append(handlers, &core.FlatStorageHandler{})
	handlers = append(handlers, &core.Executor{})

	pool := nodes.NewClientsPool(options.WatchRegistry)

	return nodes.NewRouter(pool, handlers)
}
