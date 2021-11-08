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

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/proto/tree"
)

type ctxBranchInfoKey struct{}

func WithBranchInfo(ctx context.Context, identifier string, branchInfo BranchInfo, reset ...bool) context.Context {
	//log.Logger(ctx).Error("branchInfo", zap.Any("branchInfo", branchInfo))
	value := ctx.Value(ctxBranchInfoKey{})
	var data map[string]BranchInfo
	if value != nil && len(reset) == 0 {
		data = value.(map[string]BranchInfo)
	} else {
		data = make(map[string]BranchInfo)
	}
	data[identifier] = branchInfo
	return context.WithValue(ctx, ctxBranchInfoKey{}, data)
}

func GetBranchInfo(ctx context.Context, identifier string) (BranchInfo, bool) {
	value := ctx.Value(ctxBranchInfoKey{})
	if value != nil {
		data := value.(map[string]BranchInfo)
		if info, ok := data[identifier]; ok {
			return info, true
		}
	}
	return BranchInfo{}, false
}

func AncestorsListFromContext(ctx context.Context, node *tree.Node, identifier string, p SourcesPool, orParents bool) (updatedContext context.Context, parentsList []*tree.Node, e error) {

	branchInfo, hasBranchInfo := GetBranchInfo(ctx, identifier)
	if hasBranchInfo && branchInfo.AncestorsList != nil {
		if ancestors, ok := branchInfo.AncestorsList[node.Path]; ok {
			return ctx, ancestors, nil
		}
	}
	searchFunc := BuildAncestorsList
	n := node.Clone()
	if orParents {
		n.Uuid = "" // Make sure to look by path
		searchFunc = BuildAncestorsListOrParent
	}
	parents, err := searchFunc(ctx, p.GetTreeClient(), n)
	if err != nil {
		return ctx, nil, err
	}

	if hasBranchInfo {
		if branchInfo.AncestorsList == nil {
			branchInfo.AncestorsList = make(map[string][]*tree.Node, 1)
		}
		// Make sure to detect ws_root
		for _, rootId := range branchInfo.RootUUIDs {
			for i := 0; i < len(parents); i++ {
				if parents[i].Uuid == rootId {
					cloneNode := parents[i].Clone()
					cloneNode.SetMeta(common.MetaFlagWorkspaceRoot, "true")
					parents[i] = cloneNode
				}
			}
		}
		branchInfo.AncestorsList[node.Path] = parents
		ctx = WithBranchInfo(ctx, identifier, branchInfo)
	}
	if orParents && len(parents) > 0 && parents[0].Path != n.Path {
		parents = append([]*tree.Node{n}, parents...)
	}
	return ctx, parents, nil

}

func IsFlatStorage(ctx context.Context, identifier string) bool {
	if info, ok := GetBranchInfo(ctx, identifier); ok && info.FlatStorage && !info.Binary {
		return true
	}
	return false
}
