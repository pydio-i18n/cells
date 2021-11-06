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

package abstract

import (
	"context"
	"fmt"
	"io"

	"github.com/micro/go-micro/client"
	"go.uber.org/zap"

	"github.com/pydio/cells/common/log"
	"github.com/pydio/cells/common/nodes"
	"github.com/pydio/cells/common/nodes/models"
	"github.com/pydio/cells/common/proto/tree"
)

type ContextWrapper func(ctx context.Context) (context.Context, error)

// AbstractHandler provides the simplest implementation of Client and forwards
// all calls to the Next handler
type AbstractHandler struct {
	Next        nodes.Client
	ClientsPool nodes.SourcesPool
	CtxWrapper  ContextWrapper
}

func (a *AbstractHandler) WrapContext(ctx context.Context) (context.Context, error) {
	if a.CtxWrapper != nil {
		return a.CtxWrapper(ctx)
	} else {
		return ctx, nil
	}
}

func (a *AbstractHandler) SetNextHandler(h nodes.Client) {
	a.Next = h
}

func (a *AbstractHandler) SetClientsPool(p nodes.SourcesPool) {
	a.ClientsPool = p
}

func (a *AbstractHandler) ExecuteWrapped(inputFilter nodes.NodeFilter, outputFilter nodes.NodeFilter, provider nodes.NodesCallback) error {
	wrappedIn := func(ctx context.Context, inputNode *tree.Node, identifier string) (context.Context, *tree.Node, error) {
		ctx, outputNode, err := inputFilter(ctx, inputNode, identifier)
		if err != nil {
			return ctx, inputNode, err
		}
		ctx, err = a.WrapContext(ctx)
		if err != nil {
			return ctx, inputNode, err
		}
		return ctx, outputNode, nil
	}
	return a.Next.ExecuteWrapped(wrappedIn, outputFilter, provider)
}

func (a *AbstractHandler) ReadNode(ctx context.Context, in *tree.ReadNodeRequest, opts ...client.CallOption) (*tree.ReadNodeResponse, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	log.Logger(ctx).Debug("ReadNode", zap.Any("handler", fmt.Sprintf("%T", a.Next)))
	return a.Next.ReadNode(ctx, in, opts...)
}

func (a *AbstractHandler) ListNodes(ctx context.Context, in *tree.ListNodesRequest, opts ...client.CallOption) (tree.NodeProvider_ListNodesClient, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.ListNodes(ctx, in, opts...)
}

func (a *AbstractHandler) CreateNode(ctx context.Context, in *tree.CreateNodeRequest, opts ...client.CallOption) (*tree.CreateNodeResponse, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.CreateNode(ctx, in, opts...)
}

func (a *AbstractHandler) UpdateNode(ctx context.Context, in *tree.UpdateNodeRequest, opts ...client.CallOption) (*tree.UpdateNodeResponse, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.UpdateNode(ctx, in, opts...)
}

func (a *AbstractHandler) DeleteNode(ctx context.Context, in *tree.DeleteNodeRequest, opts ...client.CallOption) (*tree.DeleteNodeResponse, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.DeleteNode(ctx, in, opts...)
}

func (a *AbstractHandler) StreamChanges(ctx context.Context, in *tree.StreamChangesRequest, opts ...client.CallOption) (tree.NodeChangesStreamer_StreamChangesClient, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.StreamChanges(ctx, in, opts...)
}

func (a *AbstractHandler) GetObject(ctx context.Context, node *tree.Node, requestData *models.GetRequestData) (io.ReadCloser, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return nil, err
	}
	return a.Next.GetObject(ctx, node, requestData)
}

func (a *AbstractHandler) PutObject(ctx context.Context, node *tree.Node, reader io.Reader, requestData *models.PutRequestData) (int64, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return 0, err
	}
	return a.Next.PutObject(ctx, node, reader, requestData)
}

func (a *AbstractHandler) CopyObject(ctx context.Context, from *tree.Node, to *tree.Node, requestData *models.CopyRequestData) (int64, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return 0, err
	}
	return a.Next.CopyObject(ctx, from, to, requestData)
}

func (a *AbstractHandler) WrappedCanApply(srcCtx context.Context, targetCtx context.Context, operation *tree.NodeChangeEvent) error {
	return a.Next.WrappedCanApply(srcCtx, targetCtx, operation)
}

// Multi part upload management

func (a *AbstractHandler) MultipartCreate(ctx context.Context, target *tree.Node, requestData *models.MultipartRequestData) (string, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return "", err
	}
	return a.Next.MultipartCreate(ctx, target, requestData)
}

func (a *AbstractHandler) MultipartPutObjectPart(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, reader io.Reader, requestData *models.PutRequestData) (models.MultipartObjectPart, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return models.MultipartObjectPart{PartNumber: partNumberMarker}, err
	}
	return a.Next.MultipartPutObjectPart(ctx, target, uploadID, partNumberMarker, reader, requestData)
}

func (a *AbstractHandler) MultipartComplete(ctx context.Context, target *tree.Node, uploadID string, uploadedParts []models.MultipartObjectPart) (models.S3ObjectInfo, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return models.S3ObjectInfo{}, err
	}
	return a.Next.MultipartComplete(ctx, target, uploadID, uploadedParts)
}

func (a *AbstractHandler) MultipartAbort(ctx context.Context, target *tree.Node, uploadID string, requestData *models.MultipartRequestData) error {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return err
	}
	return a.Next.MultipartAbort(ctx, target, uploadID, requestData)
}

func (a *AbstractHandler) MultipartList(ctx context.Context, prefix string, requestData *models.MultipartRequestData) (models.ListMultipartUploadsResult, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return models.ListMultipartUploadsResult{}, err
	}
	return a.Next.MultipartList(ctx, prefix, requestData)
}

func (a *AbstractHandler) MultipartListObjectParts(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, maxParts int) (models.ListObjectPartsResult, error) {
	ctx, err := a.WrapContext(ctx)
	if err != nil {
		return models.ListObjectPartsResult{}, err
	}
	return a.Next.MultipartListObjectParts(ctx, target, uploadID, partNumberMarker, maxParts)
}

func (a *AbstractHandler) ListNodesWithCallback(ctx context.Context, request *tree.ListNodesRequest, callback nodes.WalkFunc, ignoreCbError bool, filters ...nodes.WalkFilter) error {
	return nodes.HandlerListNodesWithCallback(a, ctx, request, callback, ignoreCbError, filters...)
}
