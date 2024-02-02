/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
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

// Package meta add persistence layer for meta data defined by the end users to enrich the nodes.
//
// Meta data might be defined by an admin and modified by normal end-users.
// Typically, to manage bookmarks or ratings.
package meta

import (
	"context"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/common/sql/resources"
	"github.com/pydio/cells/v4/idm/meta/namespace"
	"gorm.io/gorm"
)

// DAO interface
type DAO interface {
	resources.DAO

	GetNamespaceDao() namespace.DAO

	Set(ctx context.Context, meta *idm.UserMeta) (*idm.UserMeta, string, error)
	Del(ctx context.Context, meta *idm.UserMeta) (prevValue string, e error)
	Search(ctx context.Context, query sql.Enquirer) ([]*idm.UserMeta, error)
}

func NewDAO(db *gorm.DB) DAO {
	resourcesDAO := resources.NewDAO(db)
	nsDAO := namespace.NewDAO(db)

	return &sqlimpl{
		db:           db,
		resourcesDAO: resourcesDAO,
		nsDAO:        nsDAO,
	}
}
