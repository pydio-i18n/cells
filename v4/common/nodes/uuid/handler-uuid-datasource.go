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

package uuid

import (
	"context"

	"github.com/pydio/cells/v4/common/nodes/abstract"

	"github.com/micro/micro/v3/service/errors"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/proto/tree"
)

func WithDatasource() nodes.Option {
	return func(options *nodes.RouterOptions) {
		options.Wrappers = append(options.Wrappers, NewUuidDataSourceHandler())
	}
}

func NewUuidDataSourceHandler() *UuidDataSourceHandler {
	bt := &UuidDataSourceHandler{}
	bt.InputMethod = bt.updateInputBranch
	bt.OutputMethod = bt.updateOutputNode
	return bt
}

// UuidDataSourceHandler is an AbstractBranchFilter extracting datasource info based on node UUID.
type UuidDataSourceHandler struct {
	abstract.AbstractBranchFilter
}

func (h *UuidDataSourceHandler) Adapt(c nodes.Handler, options nodes.RouterOptions) nodes.Handler {
	h.Next = c
	h.ClientsPool = options.Pool
	return h
}

func (v *UuidDataSourceHandler) updateInputBranch(ctx context.Context, node *tree.Node, identifier string) (context.Context, *tree.Node, error) {

	branchInfo, ok := nodes.GetBranchInfo(ctx, identifier)
	if !ok {
		return ctx, node, errors.InternalServerError(nodes.VIEWS_LIBRARY_NAME, "Cannot find branch info for node")
	}
	if branchInfo.Client != nil {
		// DS is already set by a previous middleware, ignore.
		return ctx, node, nil
	}

	dsName := node.GetStringMeta(common.MetaNamespaceDatasourceName)
	dsPath := node.GetStringMeta(common.MetaNamespaceDatasourcePath)
	if len(dsPath) == 0 || len(dsName) == 0 {
		// Ignore this step
		return ctx, node, nil
	}
	source, e := v.ClientsPool.GetDataSourceInfo(dsName)
	if e != nil {
		return ctx, node, e
	}
	branchInfo.LoadedSource = source
	ctx = nodes.WithBranchInfo(ctx, identifier, branchInfo)

	return ctx, node, nil

}

func (v *UuidDataSourceHandler) updateOutputNode(ctx context.Context, node *tree.Node, identifier string) (context.Context, *tree.Node, error) {

	return ctx, node, nil

}
