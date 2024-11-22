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

// Package oauth provides OAuth service
package oauth

import (
	"github.com/pydio/cells/v5/common/proto/auth"
	"github.com/pydio/cells/v5/common/service"
)

var (
	PatDrivers service.StorageDrivers
)

// PatDAO interface
type PatDAO interface {
	// Load finds a corresponding, non-expired PAT based on the AccessToken.
	Load(accessToken string) (*auth.PersonalAccessToken, error)
	// Store inserts a PAT in the storage.
	Store(accessToken string, token *auth.PersonalAccessToken, update bool) error
	// Delete removes a PAT by its UUID.
	Delete(patUuid string) error
	// List lists all known PAT with optional filters.
	List(byType auth.PatType, byUser string) ([]*auth.PersonalAccessToken, error)
	// PruneExpired removes expired PAT from the storage.
	PruneExpired() (int, error)
}
