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

package oauth

import (
	"context"
	"fmt"
	"github.com/ory/hydra/v2/consent"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRange(t *testing.T) {
	Convey("Test Range of string", t, func() {
		str := rangeFromStr("http://localhost:[30000-30005]")
		So(len(str), ShouldEqual, 6)

		strFail := rangeFromStr("http://localhost:[30000-29995]")
		So(len(strFail), ShouldEqual, 1)
	})
}

func TestRegistry(t *testing.T) {
	Convey("Test Registry", t, func() {
		r := NewRegistrySQL()
		req := &consent.LoginRequest{
			ID:         "testlogin",
			ClientID:   "testclient",
			RequestURL: "testurl",
		}

		fmt.Println("And the client is ? ", req.ID, req.ClientID, req.RequestURL)

		r.ConsentManager().CreateLoginRequest(context.TODO(), req)

		resp, _ := r.ConsentManager().GetLoginRequest(context.TODO(), "testlogin")
		fmt.Println("And the client is ? ", resp.ID, resp.ClientID, resp.RequestURL)
	})
}
