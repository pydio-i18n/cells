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

	"github.com/pydio/cells/common/nodes/abstract"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/pydio/cells/common/log"
	"github.com/pydio/cells/common/nodes"
	"github.com/pydio/cells/common/nodes/acl"
	"github.com/pydio/cells/common/nodes/archive"
	"github.com/pydio/cells/common/nodes/core"
	"github.com/pydio/cells/common/nodes/encryption"
	"github.com/pydio/cells/common/nodes/path"
	"github.com/pydio/cells/common/nodes/put"
	"github.com/pydio/cells/common/nodes/version"
	"github.com/pydio/cells/common/nodes/virtual"
	"github.com/pydio/cells/common/proto/idm"
	"github.com/pydio/cells/common/proto/tree"
	"github.com/pydio/cells/common/utils/permissions"
)

// RouterEventFilter is an extended Router used mainly to filter events sent from inside to outside the application
type RouterEventFilter struct {
	nodes.Router
	rootsCache *cache.Cache
}

// NewRouterEventFilter creates a new EventFilter properly initialized
func NewRouterEventFilter(options nodes.RouterOptions) *RouterEventFilter {

	handlers := []nodes.Client{
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
		&put.PutHandler{},               // Client adding a node precreation on PUT file request
		&encryption.EncryptionHandler{}, // Client retrieve encryption materials from encryption service
		&version.VersionHandler{},
		&core.Executor{},
	)
	pool := nodes.NewClientsPool(options.WatchRegistry)
	r := nodes.NewRouter(pool, handlers)
	re := &RouterEventFilter{
		Router:     *r,
		rootsCache: cache.New(120*time.Second, 10*time.Minute),
	}
	return re

}

// WorkspaceCanSeeNode will check workspaces roots to see if a node in below one of them
func (r *RouterEventFilter) WorkspaceCanSeeNode(ctx context.Context, accessList *permissions.AccessList, workspace *idm.Workspace, node *tree.Node) (*tree.Node, bool) {
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
			log.Logger(ctx).Debug("Router Filtered node", zap.String("rootPath", parent.Path), zap.String("workspace", workspace.Label), zap.String("from", node.Path), zap.String("to", newNode.Path))
			return newNode, true
		}
	}
	return nil, false
}

// NodeIsChildOfRoot compares pathes between possible parent and child
func (r *RouterEventFilter) NodeIsChildOfRoot(ctx context.Context, node *tree.Node, rootId string) (*tree.Node, bool) {

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
func (r *RouterEventFilter) getRoot(ctx context.Context, rootId string) *tree.Node {

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
