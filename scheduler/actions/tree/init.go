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

// Package tree provides default implementation for tree related tasks.
package tree

import "github.com/pydio/cells/v5/scheduler/actions"

func init() {

	manager := actions.GetActionsManager()

	manager.Register(copyMoveActionName, func() actions.ConcreteAction {
		return &CopyMoveAction{}
	})

	manager.Register(deleteActionName, func() actions.ConcreteAction {
		return &DeleteAction{}
	})

	manager.Register(metaActionName, func() actions.ConcreteAction {
		return &MetaAction{}
	})

	manager.Register(cellsHashActionName, func() actions.ConcreteAction {
		return &CellsHashAction{}
	})

	manager.Register(datasourceAttributeActionName, func() actions.ConcreteAction {
		return &datasourceAttributeAction{}
	})

	manager.Register(middlewareMetaActionName, func() actions.ConcreteAction {
		return &middlewareMetaAction{}
	})
}
