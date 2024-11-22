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
	"github.com/pydio/cells/v5/common/proto/idm"
	"github.com/pydio/cells/v5/common/utils/merger"
)

type ACL idm.ACL

func (a *ACL) Equals(o merger.Differ) bool {
	other := o.(*ACL)

	res := a.NodeID == other.NodeID
	res = res && a.RoleID == other.RoleID
	res = res && a.WorkspaceID == other.WorkspaceID
	res = res && a.Action.Name == other.Action.Name && a.Action.Value == other.Action.Value

	return res
}

func (a *ACL) IsDeletable(m map[string]string) bool {
	return false
}

// IsMergeable tests if two user can be mergeable whose the same login name and auth source
func (a *ACL) IsMergeable(o merger.Differ) bool {
	other := o.(*ACL)

	res := a.NodeID == other.NodeID
	res = res && a.RoleID == other.RoleID
	res = res && a.WorkspaceID == other.WorkspaceID
	res = res && a.Action.Name == other.Action.Name
	return res
}

func (a *ACL) GetUniqueId() string {
	return "node" + a.NodeID + "role" + a.RoleID + "workspace" + a.WorkspaceID + "action" + a.Action.Name + a.Action.Value
}

func (a *ACL) Merge(o merger.Differ, options map[string]string) (merger.Differ, error) {
	return a, nil
}
