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
	"context"
	"strings"
	"time"

	"github.com/pydio/cells/v4/common/nodes/encryption"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/abstract"
	"github.com/pydio/cells/v4/common/nodes/acl"
	"github.com/pydio/cells/v4/common/nodes/archive"
	"github.com/pydio/cells/v4/common/nodes/core"
	"github.com/pydio/cells/v4/common/nodes/path"
	"github.com/pydio/cells/v4/common/nodes/put"
	"github.com/pydio/cells/v4/common/nodes/version"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/utils/permissions"
)

// Reverse is an extended clientImpl used mainly to filter events sent from inside to outside the application
type Reverse struct {
	nodes.Client
	rootsCache *cache.Cache
}

func ReverseClient(oo ...nodes.Option) *Reverse {
	opts := append(oo,
		nodes.WithCore(func(pool nodes.SourcesPool) nodes.Handler {
			exe := &core.Executor{}
			exe.SetClientsPool(pool)
			return exe
		}),
		acl.WithAccessList(),
		path.WithWorkspace(),
		path.WithMultipleRoots(),
		path.WithRootResolver(),
		path.WithDatasource(),
		archive.WithArchives(),
		put.WithPutInterceptor(),
		version.WithVersions(),
		encryption.WithEncryption(),
	)
	cl := newClient(opts...)
	return &Reverse{
		Client:     cl,
		rootsCache: cache.New(120*time.Second, 10*time.Minute),
	}
}

// NewRouterEventFilter creates a new EventFilter properly initialized
/*
func NewRouterEventFilter(options nodes.RouterOptions) *Reverse {

	handlers := []nodes.Handler{
		acl.NewAccessListHandler(options.AdminView),
		path.NewPathWorkspaceHandler(),
		path.NewPathMultipleRootsHandler(),
	}
	if !options.AdminView {
		handlers = append(handlers, virtual.NewVirtualNodesHandler())
	}
	handlers = append(handlers,
		path.NewWorkspaceRootResolver(),
		path.NewPathDataSourceHandler(),
		archive.NewArchiveHandler(),     // Catch "GET" request on folder.zip and create archive on-demand
		&put.PutHandler{},               // Handler adding a node precreation on PUT file request
		&encryption.EncryptionHandler{}, // Handler retrieve encryption materials from encryption service
		&version.VersionHandler{},
		&core.Executor{},
	)
	pool := nodes.NewClientsPool(options.WatchRegistry)
	r := NewRouter(pool, handlers)
	re := &Reverse{
		Client:     r,
		rootsCache: cache.New(120*time.Second, 10*time.Minute),
	}
	return re

}

*/

// WorkspaceCanSeeNode will check workspaces roots to see if a node in below one of them
func (r *Reverse) WorkspaceCanSeeNode(ctx context.Context, accessList *permissions.AccessList, workspace *idm.Workspace, node *tree.Node) (*tree.Node, bool) {
	if node == nil {
		return node, false
	}
	if tree.IgnoreNodeForOutput(ctx, node) {
		return node, false
	}
	roots := workspace.RootUUIDs
	var ancestors []*tree.Node
	var ancestorsLoaded bool
	resolver := abstract.GetVirtualNodesManager().GetResolver(r.GetClientsPool(), false)
	for _, root := range roots {
		if parent, ok := r.NodeIsChildOfRoot(ctx, node, root); ok {
			if accessList != nil {
				if !ancestorsLoaded {
					var e error
					if ancestors, e = nodes.BuildAncestorsList(ctx, r.GetClientsPool().GetTreeClient(), node); e != nil {
						log.Logger(ctx).Debug("Cannot list ancestors list for", node.Zap(), zap.Error(e))
						return node, false
					} else {
						ancestorsLoaded = true
					}
				}
				if !accessList.CanReadPath(ctx, resolver, ancestors...) {
					continue
				}
			}
			newNode := node.Clone()
			r.WrapCallback(func(inputFilter nodes.NodeFilter, outputFilter nodes.NodeFilter) error {
				branchInfo := nodes.BranchInfo{}
				branchInfo.Workspace = *workspace
				branchInfo.Root = parent
				ctx = nodes.WithBranchInfo(ctx, "in", branchInfo)
				_, newNode, _ = outputFilter(ctx, newNode, "in")
				return nil
			})
			log.Logger(ctx).Debug("clientImpl Filtered node", zap.String("rootPath", parent.Path), zap.String("workspace", workspace.Label), zap.String("from", node.Path), zap.String("to", newNode.Path))
			return newNode, true
		}
	}
	return nil, false
}

// NodeIsChildOfRoot compares pathes between possible parent and child
func (r *Reverse) NodeIsChildOfRoot(ctx context.Context, node *tree.Node, rootId string) (*tree.Node, bool) {

	vManager := abstract.GetVirtualNodesManager()
	if virtualNode, exists := vManager.ByUuid(rootId); exists {
		if resolved, e := vManager.ResolveInContext(ctx, virtualNode, r.GetClientsPool(), false); e == nil {
			//log.Logger(ctx).Info("NodeIsChildOfRoot, Comparing Pathes on resolved", zap.String("node", node.Path), zap.String("root", resolved.Path))
			return resolved, node.Path == resolved.Path || strings.HasPrefix(node.Path, strings.TrimRight(resolved.Path, "/")+"/")
		}
	}
	if root := r.getRoot(ctx, rootId); root != nil {
		//log.Logger(ctx).Info("NodeIsChildOfRoot, Comparing Pathes", zap.String("node", node.Path), zap.String("root", root.Path))
		return root, node.Path == root.Path || strings.HasPrefix(node.Path, strings.TrimRight(root.Path, "/")+"/")
	}
	return nil, false

}

// getRoot provides a loaded root node from the cache or from the treeClient
func (r *Reverse) getRoot(ctx context.Context, rootId string) *tree.Node {

	if node, ok := r.rootsCache.Get(rootId); ok {
		return node.(*tree.Node)
	}
	resp, e := r.GetClientsPool().GetTreeClient().ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{Uuid: rootId}})
	if e == nil && resp.Node != nil {
		resp.Node.Path = strings.Trim(resp.Node.Path, "/")
		r.rootsCache.Set(rootId, resp.Node.Clone(), cache.DefaultExpiration)
		return resp.Node
	}
	return nil

}
