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

package models

import (
	"context"

	"github.com/pydio/cells/v5/common/proto/idm"
	"github.com/pydio/cells/v5/common/proto/rest"
	"github.com/pydio/cells/v5/common/utils/merger"
)

type SyncShare struct {
	OwnerUser    *idm.User
	OwnerContext context.Context

	Cell           *rest.Cell
	Link           *rest.ShareLink
	LinkPassword   string
	PasswordHashed bool

	InternalData interface{}
}

func (s *SyncShare) Equals(o merger.Differ) bool {
	return false
}

func (s *SyncShare) IsDeletable(m map[string]string) bool {
	return false
}

func (s *SyncShare) IsMergeable(o merger.Differ) bool {
	return false
}

func (s *SyncShare) GetUniqueId() string {
	if s.Link != nil {
		return "LINK:" + s.Link.LinkHash
	} else if s.Cell != nil {
		return "CELL:" + s.Cell.Label
	} else {
		return "EMPTY"
	}
}

func (s *SyncShare) Merge(o merger.Differ, options map[string]string) (merger.Differ, error) {
	return s, nil
}

func (s *SyncShare) GetInternalData() interface{} {
	return s.InternalData
}
