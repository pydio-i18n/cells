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

// Package stub is a helper for testing indexation
package stub

import (
	"context"

	"github.com/pydio/cells/v4/common/proto/tree"
)

type Engine struct{}

func (s *Engine) IndexNode(context.Context, *tree.Node, bool, map[string]struct{}) error {
	return nil
}

func (s *Engine) DeleteNode(context.Context, *tree.Node) error {
	return nil
}

func (s *Engine) SearchNodes(c context.Context, queryObject *tree.Query, from int32, size int32, resultChan chan *tree.Node, facets chan *tree.SearchFacet, doneChan chan bool) error {

	resultChan <- &tree.Node{
		Uuid: "DocID1",
		Path: "/path/to/node.txt",
	}

	doneChan <- true

	return nil
}

func (s *Engine) Close() error {
	return nil
}

func (s *Engine) ClearIndex(ctx context.Context) error {
	return nil
}
