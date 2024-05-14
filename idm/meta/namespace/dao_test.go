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

package namespace

import (
	"context"
	"github.com/pydio/cells/v4/common/sql"

	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/proto/idm"
	service "github.com/pydio/cells/v4/common/proto/service"
	"github.com/pydio/cells/v4/common/runtime"
	_ "github.com/pydio/cells/v4/common/utils/cache/gocache"
	"github.com/pydio/cells/v4/common/utils/configx"
	json "github.com/pydio/cells/v4/common/utils/jsonx"
	"github.com/spf13/viper"
)

var (
	mockDAO DAO
)

func TestMain(m *testing.M) {
	v := viper.New()
	v.SetDefault(runtime.KeyCache, "pm://")
	v.SetDefault(runtime.KeyShortCache, "pm://")
	runtime.SetRuntime(v)

	var options = configx.New()
	ctx := context.Background()
	if d, e := dao.InitDAO(ctx, sql.SqliteDriver, sql.SharedMemDSN, "test", NewDAO, options); e != nil {
		panic(e)
	} else {
		mockDAO = d.(DAO)
	}
	m.Run()
}

func TestCrud(t *testing.T) {

	Convey("Create Meta Namespace", t, func() {
		// Insert a meta
		err := mockDAO.Add(&idm.UserMetaNamespace{
			Namespace:      "namespace",
			Label:          "label",
			Order:          1,
			JsonDefinition: "{\"test\":\"value\"}",
		})
		So(err, ShouldBeNil)

		// List meta
		result, er := mockDAO.List()
		So(er, ShouldBeNil)
		So(result, ShouldHaveLength, 2) // 2 because DAO automatically adds the Bookmarks namespace
		So(result["namespace"].Order, ShouldEqual, 1)

		jsonDef := result["namespace"].JsonDefinition
		var def map[string]string
		er = json.Unmarshal([]byte(jsonDef), &def)
		So(er, ShouldBeNil)
		So(def["test"], ShouldEqual, "value")

		e := mockDAO.Del(&idm.UserMetaNamespace{Namespace: "namespace"})
		So(e, ShouldBeNil)

		// List meta for the node
		result2, er := mockDAO.List()
		So(er, ShouldBeNil)
		So(result2, ShouldHaveLength, 1)
	})

}

func TestResourceRules(t *testing.T) {

	Convey("Test Add Rule", t, func() {

		err := mockDAO.AddPolicy("resource-id", &service.ResourcePolicy{Action: service.ResourcePolicyAction_READ, Subject: "subject1"})
		So(err, ShouldBeNil)

	})

	Convey("Select Rules", t, func() {

		rules, err := mockDAO.GetPoliciesForResource("resource-id")
		So(rules, ShouldHaveLength, 1)
		So(err, ShouldBeNil)

	})

	Convey("Delete Rules", t, func() {

		err := mockDAO.DeletePoliciesForResource("resource-id")
		So(err, ShouldBeNil)

		rules, err := mockDAO.GetPoliciesForResource("resource-id")
		So(rules, ShouldHaveLength, 0)
		So(err, ShouldBeNil)

	})

	Convey("Delete Rules For Action", t, func() {

		mockDAO.AddPolicy("resource-id", &service.ResourcePolicy{Action: service.ResourcePolicyAction_READ, Subject: "subject1"})
		mockDAO.AddPolicy("resource-id", &service.ResourcePolicy{Action: service.ResourcePolicyAction_WRITE, Subject: "subject1"})

		rules, err := mockDAO.GetPoliciesForResource("resource-id")
		So(rules, ShouldHaveLength, 2)

		err = mockDAO.DeletePoliciesForResourceAndAction("resource-id", service.ResourcePolicyAction_READ)
		So(err, ShouldBeNil)

		rules, err = mockDAO.GetPoliciesForResource("resource-id")
		So(rules, ShouldHaveLength, 1)
		So(err, ShouldBeNil)

	})

}
