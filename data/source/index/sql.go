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

package index

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/common/sql/index"
	"github.com/pydio/cells/v4/common/utils/configx"
	"github.com/pydio/cells/v4/common/utils/mtree"
)

var (
	queries = map[string]interface{}{}
)

type sqlimpl struct {
	*sql.Handler

	*index.IndexSQL
}

// Init handler for the SQL DAO
func (s *sqlimpl) Init(ctx context.Context, options configx.Values) error {

	// super
	s.DAO.Init(ctx, options)

	// Preparing the index
	s.IndexSQL = index.NewDAO(s.Handler, "ROOT").(*index.IndexSQL)
	if err := s.IndexSQL.Init(ctx, options); err != nil {
		return err
	}

	log.Logger(context.Background()).Debug("Finished IndexSQL Init")

	// Preparing the db statements
	if options.Val("prepare").Default(true).Bool() {
		for key, query := range queries {
			if err := s.Prepare(key, query); err != nil {
				return err
			}
		}
	}

	log.Logger(context.Background()).Debug("Local sql Prepares")

	if _, err := s.IndexSQL.GetNode(mtree.NewMPath(1)); err != nil {
		log.Logger(context.Background()).Info("Creating root node in index ")
		treeNode := mtree.NewTreeNode()
		treeNode.SetType(tree.NodeType_COLLECTION)
		treeNode.UpdateUuid("ROOT")
		treeNode.SetMPath(1)
		treeNode.Level = 1
		treeNode.UpdateMTime(time.Now().Unix())
		s.IndexSQL.AddNode(treeNode)
	}

	return nil
}

// NewNode utils
func NewNode(treeNode tree.N, path mtree.MPath, filenames []string) *mtree.TreeNode {

	node := mtree.NewTreeNode()
	if treeNode == nil {
		treeNode = tree.LightNode(tree.NodeType_UNKNOWN, "", "", "", 0, 0, 0)
	}
	node.N = treeNode

	node.SetMPath(path...)
	node.SetName(filenames[len(filenames)-1])

	node.UpdatePath(strings.Join(filenames, string(os.PathSeparator)))

	return node
}
