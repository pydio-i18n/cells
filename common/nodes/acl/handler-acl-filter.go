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

package acl

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/abstract"
	"github.com/pydio/cells/v4/common/nodes/models"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/service/errors"
	"github.com/pydio/cells/v4/common/utils/permissions"
)

var pathNotReadable = errors.Forbidden("path.not.readable", "path is not readable")
var pathNotWriteable = errors.Forbidden("path.not.writeable", "path is not writeable")

func WithFilter() nodes.Option {
	return func(options *nodes.RouterOptions) {
		if !options.AdminView {
			options.Wrappers = append(options.Wrappers, &FilterHandler{})
		}
	}
}

// FilterHandler checks for read/write permissions depending on the call using the context AccessList.
type FilterHandler struct {
	abstract.Handler
}

func (a *FilterHandler) Adapt(h nodes.Handler, options nodes.RouterOptions) nodes.Handler {
	a.AdaptOptions(h, options)
	return a
}

func (a *FilterHandler) skipContext(ctx context.Context, identifier ...string) bool {
	if nodes.HasSkipAclCheck(ctx) {
		return true
	}
	id := "in"
	if len(identifier) > 0 {
		id = identifier[0]
	}
	bI, ok := nodes.GetBranchInfo(ctx, id)
	return ok && (bI.Binary || bI.TransparentBinary)
}

// ReadNode checks if node is readable and forward to next middleware.
func (a *FilterHandler) ReadNode(ctx context.Context, in *tree.ReadNodeRequest, opts ...grpc.CallOption) (*tree.ReadNodeResponse, error) {
	if a.skipContext(ctx) {
		return a.Next.ReadNode(ctx, in, opts...)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)

	// First load ancestors or grab them from BranchInfo
	ctx, parents, err := nodes.AncestorsListFromContext(ctx, in.Node, "in", a.ClientsPool, false)
	if err != nil {
		return nil, a.recheckParents(ctx, err, in.Node, true, false)
	}
	if !accessList.CanRead(ctx, parents...) && !accessList.CanWrite(ctx, parents...) {
		return nil, pathNotReadable
	}
	checkDl := in.Node.HasMetaKey(nodes.MetaAclCheckDownload)
	checkSync := in.Node.HasMetaKey(nodes.MetaAclCheckSyncable)
	response, err := a.Next.ReadNode(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	if accessList.CanRead(ctx, parents...) && !accessList.CanWrite(ctx, parents...) {
		n := response.Node.Clone()
		n.MustSetMeta(common.MetaFlagReadonly, "true")
		response.Node = n
	}
	updatedParents := append([]*tree.Node{response.GetNode()}, parents[1:]...)
	if checkDl && accessList.HasExplicitDeny(ctx, permissions.FlagDownload, updatedParents...) {
		return nil, errors.Forbidden("download.forbidden", "Node cannot be downloaded")
	}
	if checkSync && accessList.HasExplicitDeny(ctx, permissions.FlagSync, updatedParents...) {
		n := response.Node.Clone()
		n.MustSetMeta(common.MetaFlagWorkspaceSyncable, false)
		response.Node = n
	}
	return response, err
}

// ListNodes filters list results with ACLs permissions
func (a *FilterHandler) ListNodes(ctx context.Context, in *tree.ListNodesRequest, opts ...grpc.CallOption) (streamer tree.NodeProvider_ListNodesClient, e error) {
	if a.skipContext(ctx) {
		return a.Next.ListNodes(ctx, in, opts...)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	// First load ancestors or grab them from BranchInfo
	ctx, parents, err := nodes.AncestorsListFromContext(ctx, in.Node, "in", a.ClientsPool, false)
	if err != nil {
		return nil, a.recheckParents(ctx, err, in.Node, true, false)
	}

	if !accessList.CanRead(ctx, parents...) {
		return nil, errors.Forbidden("node.not.readable", "Node is not readable")
	}

	stream, err := a.Next.ListNodes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	s := nodes.NewWrappingStreamer(stream.Context())
	go func() {
		defer s.CloseSend()
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					s.SendError(err)
				}
				break
			}
			if resp == nil {
				continue
			}
			// FILTER OUT NON READABLE NODES
			newBranch := []*tree.Node{resp.Node}
			newBranch = append(newBranch, parents...)
			if !accessList.CanRead(ctx, newBranch...) {
				continue
			}
			if accessList.CanRead(ctx, newBranch...) && !accessList.CanWrite(ctx, newBranch...) {
				n := resp.Node.Clone()
				n.MustSetMeta(common.MetaFlagReadonly, "true")
				resp.Node = n
			}
			s.Send(resp)
		}
	}()

	return s, nil
}

func (a *FilterHandler) CreateNode(ctx context.Context, in *tree.CreateNodeRequest, opts ...grpc.CallOption) (*tree.CreateNodeResponse, error) {
	if a.skipContext(ctx) {
		return a.Next.CreateNode(ctx, in, opts...)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	ctx, toParents, err := nodes.AncestorsListFromContext(ctx, in.Node, "in", a.ClientsPool, true)
	if err != nil {
		return nil, err
	}
	if !accessList.CanWrite(ctx, toParents...) {
		return nil, pathNotWriteable
	}
	return a.Next.CreateNode(ctx, in, opts...)
}

func (a *FilterHandler) UpdateNode(ctx context.Context, in *tree.UpdateNodeRequest, opts ...grpc.CallOption) (*tree.UpdateNodeResponse, error) {
	if a.skipContext(ctx) {
		return a.Next.UpdateNode(ctx, in, opts...)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	ctx, fromParents, err := nodes.AncestorsListFromContext(ctx, in.From, "from", a.ClientsPool, false)
	if err != nil {
		return nil, a.recheckParents(ctx, err, in.From, true, false)
	}
	if !accessList.CanRead(ctx, fromParents...) {
		return nil, pathNotReadable
	}
	ctx, toParents, err := nodes.AncestorsListFromContext(ctx, in.To, "to", a.ClientsPool, true)
	if err != nil {
		return nil, err
	}
	if !accessList.CanWrite(ctx, toParents...) {
		return nil, pathNotWriteable
	}
	return a.Next.UpdateNode(ctx, in, opts...)
}

func (a *FilterHandler) DeleteNode(ctx context.Context, in *tree.DeleteNodeRequest, opts ...grpc.CallOption) (*tree.DeleteNodeResponse, error) {
	if a.skipContext(ctx) {
		return a.Next.DeleteNode(ctx, in, opts...)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	ctx, delParents, err := nodes.AncestorsListFromContext(ctx, in.Node, "in", a.ClientsPool, false)
	if err != nil {
		return nil, a.recheckParents(ctx, err, in.Node, true, false)
	}
	if !accessList.CanWrite(ctx, delParents...) {
		return nil, pathNotWriteable
	}
	if accessList.HasExplicitDeny(ctx, permissions.FlagDelete, delParents...) {
		return nil, errors.Forbidden("delete.forbidden", "Node cannot be deleted")
	}
	return a.Next.DeleteNode(ctx, in, opts...)
}

func (a *FilterHandler) GetObject(ctx context.Context, node *tree.Node, requestData *models.GetRequestData) (io.ReadCloser, error) {
	if a.skipContext(ctx) {
		return a.Next.GetObject(ctx, node, requestData)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	// First load ancestors or grab them from BranchInfo
	ctx, parents, err := nodes.AncestorsListFromContext(ctx, node, "in", a.ClientsPool, false)
	if err != nil {
		return nil, a.recheckParents(ctx, err, node, true, false)
	}
	if !accessList.CanRead(ctx, parents...) {
		return nil, pathNotReadable
	}
	if accessList.HasExplicitDeny(ctx, permissions.FlagDownload, parents...) {
		return nil, errors.Forbidden("download.forbidden", "Node is not downloadable")
	}
	return a.Next.GetObject(ctx, node, requestData)
}

func (a *FilterHandler) PutObject(ctx context.Context, node *tree.Node, reader io.Reader, requestData *models.PutRequestData) (models.ObjectInfo, error) {
	if a.skipContext(ctx) {
		return a.Next.PutObject(ctx, node, reader, requestData)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	// First load ancestors or grab them from BranchInfo
	checkNode := node.Clone()
	checkNode.Type = tree.NodeType_LEAF
	checkNode.Size = requestData.Size
	ctx, parents, err := nodes.AncestorsListFromContext(ctx, checkNode, "in", a.ClientsPool, true)
	if err != nil {
		return models.ObjectInfo{}, err
	}
	if !accessList.CanWrite(ctx, parents...) {
		return models.ObjectInfo{}, pathNotWriteable
	}
	if accessList.HasExplicitDeny(ctx, permissions.FlagUpload, parents...) {
		return models.ObjectInfo{}, errors.Forbidden("upload.forbidden", "Parents have upload explicitly disabled")
	}
	return a.Next.PutObject(ctx, node, reader, requestData)
}

func (a *FilterHandler) MultipartCreate(ctx context.Context, node *tree.Node, requestData *models.MultipartRequestData) (string, error) {
	if a.skipContext(ctx) {
		return a.Next.MultipartCreate(ctx, node, requestData)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	// First load ancestors or grab them from BranchInfo
	ctx, parents, err := nodes.AncestorsListFromContext(ctx, node, "in", a.ClientsPool, true)
	if err != nil {
		return "", err
	}
	if !accessList.CanWrite(ctx, parents...) {
		return "", pathNotWriteable
	}
	if accessList.HasExplicitDeny(ctx, permissions.FlagUpload, parents...) {
		return "", errors.Forbidden("upload.forbidden", "Parents have upload explicitly disabled")
	}
	return a.Next.MultipartCreate(ctx, node, requestData)
}

func (a *FilterHandler) CopyObject(ctx context.Context, from *tree.Node, to *tree.Node, requestData *models.CopyRequestData) (models.ObjectInfo, error) {
	if a.skipContext(ctx) {
		return a.Next.CopyObject(ctx, from, to, requestData)
	}
	accessList := ctx.Value(ctxUserAccessListKey{}).(*permissions.AccessList)
	ctx, fromParents, err := nodes.AncestorsListFromContext(ctx, from, "from", a.ClientsPool, false)
	if err != nil {
		return models.ObjectInfo{}, a.recheckParents(ctx, err, from, true, false)
	}
	if !accessList.CanRead(ctx, fromParents...) {
		return models.ObjectInfo{}, pathNotReadable
	}
	ctx, toParents, err := nodes.AncestorsListFromContext(ctx, to, "to", a.ClientsPool, true)
	if err != nil {
		return models.ObjectInfo{}, err
	}
	if !accessList.CanWrite(ctx, toParents...) {
		return models.ObjectInfo{}, pathNotWriteable
	}
	if accessList.HasExplicitDeny(ctx, permissions.FlagUpload, toParents...) {
		return models.ObjectInfo{}, errors.Forbidden("upload.forbidden", "Parents have upload explicitly disabled")
	}
	fullTargets := append(toParents, to)
	if accessList.HasExplicitDeny(ctx, permissions.FlagDownload, fromParents...) && !accessList.HasExplicitDeny(ctx, permissions.FlagDownload, fullTargets...) {
		return models.ObjectInfo{}, errors.Forbidden("upload.forbidden", "Source has download explicitly disabled and target does not")
	}
	return a.Next.CopyObject(ctx, from, to, requestData)
}

func (a *FilterHandler) WrappedCanApply(srcCtx context.Context, targetCtx context.Context, operation *tree.NodeChangeEvent) error {

	var rwErr error
	switch operation.GetType() {
	case tree.NodeChangeEvent_UPDATE_CONTENT:

		rwErr = a.checkPerm(targetCtx, operation.GetTarget(), "in", true, false, true, permissions.FlagUpload)

	case tree.NodeChangeEvent_CREATE:

		rwErr = a.checkPerm(targetCtx, operation.GetTarget(), "in", true, false, true)

	case tree.NodeChangeEvent_DELETE:

		rwErr = a.checkPerm(srcCtx, operation.GetSource(), "in", false, false, true, permissions.FlagDelete)

	case tree.NodeChangeEvent_UPDATE_PATH:

		if rwErr = a.checkPerm(srcCtx, operation.GetSource(), "from", false, true, true); rwErr != nil {
			return rwErr
		}
		// For delete operations, ignore write permissions as recycle can be outside of authorized paths
		if operation.GetTarget().GetStringMeta(common.RecycleBinName) != "true" {
			rwErr = a.checkPerm(targetCtx, operation.GetTarget(), "to", true, false, true)
		}

	case tree.NodeChangeEvent_READ:

		if operation.GetSource().HasMetaKey(nodes.MetaAclCheckDownload) {
			rwErr = a.checkPerm(srcCtx, operation.GetSource(), "in", false, true, false, permissions.FlagDownload)
		} else {
			rwErr = a.checkPerm(srcCtx, operation.GetSource(), "in", false, true, false)
		}

	}
	if rwErr != nil {
		return rwErr
	}
	return a.Next.WrappedCanApply(srcCtx, targetCtx, operation)
}

func (a *FilterHandler) checkPerm(c context.Context, node *tree.Node, identifier string, orParents bool, read bool, write bool, explicitFlags ...permissions.BitmaskFlag) error {

	val := c.Value(ctxUserAccessListKey{})
	if val == nil {
		return fmt.Errorf("cannot find accessList in context for checking permissions")
	}
	accessList := val.(*permissions.AccessList)
	ctx, parents, err := nodes.AncestorsListFromContext(c, node, identifier, a.ClientsPool, orParents)
	if err != nil {
		return a.recheckParents(c, err, node, read, write)
	}
	if read && !accessList.CanRead(ctx, parents...) {
		return pathNotReadable
	}
	if write && !accessList.CanWrite(ctx, parents...) {
		return pathNotWriteable
	}
	if len(explicitFlags) > 0 && accessList.HasExplicitDeny(ctx, explicitFlags[0], parents...) {
		return errors.Forbidden("explicit.deny", "path has explicit denies for flag "+permissions.FlagsToNames[explicitFlags[0]])
	}
	return nil

}

func (a *FilterHandler) recheckParents(c context.Context, originalError error, node *tree.Node, read, write bool) error {

	if errors.FromError(originalError).Code != 404 {
		return originalError
	}

	val := c.Value(ctxUserAccessListKey{})
	if val == nil {
		return fmt.Errorf("cannot find accessList in context for checking permissions")
	}
	accessList := val.(*permissions.AccessList)

	parents, e := nodes.BuildAncestorsListOrParent(c, a.ClientsPool.GetTreeClient(), node)
	if e != nil {
		return e
	}

	if read && !accessList.CanRead(c, parents...) {
		return pathNotReadable
	}
	if write && !accessList.CanWrite(c, parents...) {
		return pathNotWriteable
	}

	return originalError
}
