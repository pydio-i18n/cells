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

package nodes

import (
	"context"
	"fmt"
	"io"

	"github.com/micro/go-micro/client"

	"github.com/pydio/cells/common/nodes/models"
	"github.com/pydio/cells/common/proto/tree"
)

// NewRouter creates and configures a new router with given ClientsPool and Handlers.
func NewRouter(pool SourcesPool, handlers []Client) *Router {
	r := &Router{
		handlers: handlers,
		pool:     pool,
	}
	r.initHandlers()
	return r
}

type Router struct {
	handlers []Client
	pool     SourcesPool
}

func (v *Router) initHandlers() {
	for i, h := range v.handlers {
		if i < len(v.handlers)-1 {
			next := v.handlers[i+1]
			h.SetNextHandler(next)
		}
		h.SetClientsPool(v.pool)
	}
}

func (v *Router) WrapCallback(provider NodesCallback) error {
	return v.ExecuteWrapped(nil, nil, provider)
}

func (v *Router) BranchInfoForNode(ctx context.Context, node *tree.Node) (branch BranchInfo, err error) {
	err = v.WrapCallback(func(inputFilter NodeFilter, outputFilter NodeFilter) error {
		updatedCtx, _, er := inputFilter(ctx, node, "in")
		if er != nil {
			return er
		}
		if dsInfo, o := GetBranchInfo(updatedCtx, "in"); o {
			branch = dsInfo
		} else {
			return fmt.Errorf("cannot find branch info for node " + node.GetPath())
		}
		return nil
	})
	return
}

func (v *Router) ExecuteWrapped(_ NodeFilter, _ NodeFilter, provider NodesCallback) error {
	identity := func(ctx context.Context, inputNode *tree.Node, identifier string) (context.Context, *tree.Node, error) {
		return ctx, inputNode, nil
	}
	return v.handlers[0].ExecuteWrapped(identity, identity, provider)
}

func (v *Router) ReadNode(ctx context.Context, in *tree.ReadNodeRequest, opts ...client.CallOption) (*tree.ReadNodeResponse, error) {
	h := v.handlers[0]

	return h.ReadNode(ctx, in, opts...)
}

func (v *Router) ListNodes(ctx context.Context, in *tree.ListNodesRequest, opts ...client.CallOption) (tree.NodeProvider_ListNodesClient, error) {
	h := v.handlers[0]
	return h.ListNodes(ctx, in, opts...)
}

func (v *Router) CreateNode(ctx context.Context, in *tree.CreateNodeRequest, opts ...client.CallOption) (*tree.CreateNodeResponse, error) {
	h := v.handlers[0]
	return h.CreateNode(ctx, in, opts...)
}

func (v *Router) UpdateNode(ctx context.Context, in *tree.UpdateNodeRequest, opts ...client.CallOption) (*tree.UpdateNodeResponse, error) {
	h := v.handlers[0]
	return h.UpdateNode(ctx, in, opts...)
}

func (v *Router) DeleteNode(ctx context.Context, in *tree.DeleteNodeRequest, opts ...client.CallOption) (*tree.DeleteNodeResponse, error) {
	h := v.handlers[0]
	return h.DeleteNode(ctx, in, opts...)
}

func (v *Router) GetObject(ctx context.Context, node *tree.Node, requestData *models.GetRequestData) (io.ReadCloser, error) {
	h := v.handlers[0]
	return h.GetObject(ctx, node, requestData)
}

func (v *Router) PutObject(ctx context.Context, node *tree.Node, reader io.Reader, requestData *models.PutRequestData) (int64, error) {
	h := v.handlers[0]
	return h.PutObject(ctx, node, reader, requestData)
}

func (v *Router) CopyObject(ctx context.Context, from *tree.Node, to *tree.Node, requestData *models.CopyRequestData) (int64, error) {
	h := v.handlers[0]
	return h.CopyObject(ctx, from, to, requestData)
}

func (v *Router) MultipartCreate(ctx context.Context, target *tree.Node, requestData *models.MultipartRequestData) (string, error) {
	return v.handlers[0].MultipartCreate(ctx, target, requestData)
}

func (v *Router) MultipartPutObjectPart(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, reader io.Reader, requestData *models.PutRequestData) (models.MultipartObjectPart, error) {
	return v.handlers[0].MultipartPutObjectPart(ctx, target, uploadID, partNumberMarker, reader, requestData)
}

func (v *Router) MultipartList(ctx context.Context, prefix string, requestData *models.MultipartRequestData) (models.ListMultipartUploadsResult, error) {
	return v.handlers[0].MultipartList(ctx, prefix, requestData)
}

func (v *Router) MultipartAbort(ctx context.Context, target *tree.Node, uploadID string, requestData *models.MultipartRequestData) error {
	return v.handlers[0].MultipartAbort(ctx, target, uploadID, requestData)
}

func (v *Router) MultipartComplete(ctx context.Context, target *tree.Node, uploadID string, uploadedParts []models.MultipartObjectPart) (models.S3ObjectInfo, error) {
	return v.handlers[0].MultipartComplete(ctx, target, uploadID, uploadedParts)
}

func (v *Router) MultipartListObjectParts(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, maxParts int) (models.ListObjectPartsResult, error) {
	return v.handlers[0].MultipartListObjectParts(ctx, target, uploadID, partNumberMarker, maxParts)
}

func (v *Router) StreamChanges(ctx context.Context, in *tree.StreamChangesRequest, opts ...client.CallOption) (tree.NodeChangesStreamer_StreamChangesClient, error) {
	return v.handlers[0].StreamChanges(ctx, in, opts...)
}

func (v *Router) WrappedCanApply(srcCtx context.Context, targetCtx context.Context, operation *tree.NodeChangeEvent) error {
	return v.handlers[0].WrappedCanApply(srcCtx, targetCtx, operation)
}

func (v *Router) CanApply(ctx context.Context, operation *tree.NodeChangeEvent) (*tree.NodeChangeEvent, error) {
	var innerOperation *tree.NodeChangeEvent
	e := v.WrapCallback(func(inputFilter NodeFilter, outputFilter NodeFilter) error {
		var sourceNode, targetNode *tree.Node
		var sourceCtx, targetCtx context.Context
		switch operation.Type {
		case tree.NodeChangeEvent_CREATE, tree.NodeChangeEvent_UPDATE_CONTENT:
			targetNode = operation.Target
		case tree.NodeChangeEvent_READ, tree.NodeChangeEvent_DELETE:
			sourceNode = operation.Source
		case tree.NodeChangeEvent_UPDATE_PATH:
			targetNode = operation.Target
			sourceNode = operation.Source
		}
		var e error
		if targetNode != nil {
			targetCtx, targetNode, e = inputFilter(ctx, targetNode, "in")
			if e != nil {
				return e
			}
		}
		if sourceNode != nil {
			sourceCtx, sourceNode, e = inputFilter(ctx, targetNode, "in")
			if e != nil {
				return e
			}
		}
		innerOperation = &tree.NodeChangeEvent{Type: operation.Type, Source: sourceNode, Target: targetNode}
		return v.WrappedCanApply(sourceCtx, targetCtx, &tree.NodeChangeEvent{Type: operation.Type, Source: sourceNode, Target: targetNode})
	})
	return innerOperation, e
}

// To respect Client interface

func (v *Router) SetNextHandler(Client)      {}
func (v *Router) SetClientsPool(SourcesPool) {}

// GetExecutor uses the very last handler (Executor) to send a request with a previously filled context.
func (v *Router) GetExecutor() Client {
	return v.handlers[len(v.handlers)-1]
}

// GetClientsPool is specific to Router
func (v *Router) GetClientsPool() SourcesPool {
	return v.pool
}
