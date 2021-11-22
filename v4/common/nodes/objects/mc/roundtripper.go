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

package mc

import (
	"fmt"
	"net/http"

	minio "github.com/minio/minio-go/v7"

	"github.com/pydio/cells/v4/common"
)

func newRoundTripper(secure bool) (http.RoundTripper, error) {
	if def, e := minio.DefaultTransport(secure); e != nil {
		return nil, e
	} else {
		return &roundTripper{w: def}, nil
	}
}

type roundTripper struct {
	w http.RoundTripper
}

func (r *roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()
	if v := ctx.Value(common.PydioContextUserKey); v != nil {
		fmt.Println("Round Trip Found Context User Key", v.(string))
	} else {
		fmt.Println("Key Not found in context")
	}
	return r.w.RoundTrip(request)
}
