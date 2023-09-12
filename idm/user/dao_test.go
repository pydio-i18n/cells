/*
 * Copyright (c) 2021. Abstrium SAS <team (at) pydio.com>
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

package user

import (
	"context"
	"github.com/pydio/cells/v4/common/utils/mtree"
	"log"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/dao/sqlite"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/service"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service/errors"
	"github.com/pydio/cells/v4/common/sql"
	_ "github.com/pydio/cells/v4/common/utils/cache/gocache"
	"github.com/pydio/cells/v4/common/utils/configx"
)

var (
	mockDAO DAO

	wg sync.WaitGroup
)

type server struct{}

func TestMain(m *testing.M) {
	v := viper.New()
	v.SetDefault(runtime.KeyCache, "pm://")
	v.SetDefault(runtime.KeyShortCache, "pm://")
	runtime.SetRuntime(v)

	var options = configx.New()
	ctx := context.Background()
	if d, e := dao.InitDAO(ctx, sqlite.Driver, sqlite.SharedMemDSN, "idm_user", NewDAO, options); e != nil {
		panic(e)
	} else {
		mockDAO = d.(DAO)
	}

	m.Run()
	wg.Wait()
}

func TestQueryBuilder(t *testing.T) {

	sqliteDao := mockDAO.(*sqlimpl)
	converter := &queryConverter{
		treeDao: sqliteDao.IndexSQL,
	}

	Convey("Query Builder", t, func() {

		singleQ1, singleQ2 := new(idm.UserSingleQuery), new(idm.UserSingleQuery)

		singleQ1.Login = "user1"
		singleQ1.Password = "passwordUser1"

		singleQ2.Login = "user2"
		singleQ2.Password = "passwordUser2"

		singleQ1Any, err := anypb.New(singleQ1)
		So(err, ShouldBeNil)

		singleQ2Any, err := anypb.New(singleQ2)
		So(err, ShouldBeNil)

		var singleQueries []*anypb.Any
		singleQueries = append(singleQueries, singleQ1Any)
		singleQueries = append(singleQueries, singleQ2Any)

		simpleQuery := &service.Query{
			SubQueries: singleQueries,
			Operation:  service.OperationType_OR,
			Offset:     0,
			Limit:      10,
		}

		s := sql.NewQueryBuilder(simpleQuery, converter).Expression("sqlite3")
		So(s, ShouldNotBeNil)

	})

	Convey("Query Builder with join fields", t, func() {

		_, _, e := mockDAO.Add(&idm.User{
			Login:     "username",
			Password:  "xxxxxxx",
			GroupPath: "/path/to/group",
		})
		So(e, ShouldBeNil)

		singleQ1, singleQ2 := new(idm.UserSingleQuery), new(idm.UserSingleQuery)
		singleQ1.GroupPath = "/path/to/group"
		singleQ1.HasRole = "a_role_name"

		singleQ2.AttributeName = idm.UserAttrHidden
		singleQ2.AttributeAnyValue = true
		//		singleQ2.Not = true

		singleQ1Any, err := anypb.New(singleQ1)
		So(err, ShouldBeNil)

		singleQ2Any, err := anypb.New(singleQ2)
		So(err, ShouldBeNil)

		var singleQueries []*anypb.Any
		singleQueries = append(singleQueries, singleQ1Any)
		singleQueries = append(singleQueries, singleQ2Any)

		simpleQuery := &service.Query{
			SubQueries: singleQueries,
			Operation:  service.OperationType_AND,
			Offset:     0,
			Limit:      10,
		}

		s := sql.NewQueryBuilder(simpleQuery, converter).Expression("sqlite")
		So(s, ShouldNotBeNil)

	})

	Convey("Test DAO", t, func() {

		_, _, fail := mockDAO.Add(map[string]string{})
		So(fail, ShouldNotBeNil)

		_, _, err := mockDAO.Add(&idm.User{
			Login:     "username",
			Password:  "xxxxxxx",
			GroupPath: "/path/to/group",
			Attributes: map[string]string{
				idm.UserAttrDisplayName: "John Doe",
				idm.UserAttrHidden:      "false",
				"active":                "true",
			},
			Roles: []*idm.Role{
				{Uuid: "1", Label: "Role1"},
				{Uuid: "2", Label: "Role2"},
			},
		})

		So(err, ShouldBeNil)

		{
			users := new([]interface{})
			e := mockDAO.Search(&service.Query{Limit: -1}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 5)
		}

		{
			res, e := mockDAO.Count(&service.Query{Limit: -1})
			So(e, ShouldBeNil)
			So(res, ShouldEqual, 5)
		}

		{
			users := new([]interface{})
			e := mockDAO.Search(&service.Query{Offset: 1, Limit: 2}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 2)
		}

		{
			users := new([]interface{})
			e := mockDAO.Search(&service.Query{Offset: 4, Limit: 10}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
		}

		{
			u, e := mockDAO.Bind("username", "xxxxxxx")
			So(e, ShouldBeNil)
			So(u, ShouldNotBeNil)
		}

		{
			u, e := mockDAO.Bind("usernameXX", "xxxxxxx")
			So(u, ShouldBeNil)
			So(e, ShouldNotBeNil)
			So(errors.FromError(e).Code, ShouldEqual, 404)
		}

		{
			u, e := mockDAO.Bind("username", "xxxxxxxYY")
			So(u, ShouldBeNil)
			So(e, ShouldNotBeNil)
			So(errors.FromError(e).Code, ShouldEqual, 403)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				Login: "user1",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 0)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				Login: "username",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				Login:    "username",
				NodeType: idm.NodeType_USER,
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				Login:    "username",
				NodeType: idm.NodeType_GROUP,
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 0)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				GroupPath: "/path/to/group",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
		}

		_, _, err2 := mockDAO.Add(&idm.User{
			IsGroup:   true,
			GroupPath: "/path/to/anotherGroup",
			Attributes: map[string]string{
				"displayName": "Group Display Name",
			},
		})

		So(err2, ShouldBeNil)

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				FullPath: "/path/to/group",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
			object := (*users)[0]
			group, ok := object.(*idm.User)
			So(ok, ShouldBeTrue)
			So(group.GroupLabel, ShouldEqual, "group")
			So(group.GroupPath, ShouldEqual, "/path/to/")
			So(group.IsGroup, ShouldBeTrue)

		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				FullPath: "/path/to/anotherGroup",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
			object := (*users)[0]
			group, ok := object.(*idm.User)
			So(ok, ShouldBeTrue)
			So(group.GroupLabel, ShouldEqual, "anotherGroup")
			So(group.GroupPath, ShouldEqual, "/path/to/")
			So(group.IsGroup, ShouldBeTrue)
			So(group.Attributes, ShouldResemble, map[string]string{"displayName": "Group Display Name"})

		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				AttributeName:  "displayName",
				AttributeValue: "John*",
			}
			userQueryAny, _ := anypb.New(userQuery)
			userQuery2 := &idm.UserSingleQuery{
				AttributeName:  "active",
				AttributeValue: "true",
			}
			userQueryAny2, _ := anypb.New(userQuery2)
			userQuery3 := &idm.UserSingleQuery{
				AttributeName:  idm.UserAttrHidden,
				AttributeValue: "false",
			}
			userQueryAny3, _ := anypb.New(userQuery3)

			total, e1 := mockDAO.Count(&service.Query{
				SubQueries: []*anypb.Any{
					userQueryAny,
					userQueryAny2,
					userQueryAny3,
				},
				Operation: service.OperationType_AND,
			})
			So(e1, ShouldBeNil)
			So(total, ShouldEqual, 1)

			e := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{
					userQueryAny,
					userQueryAny2,
					userQueryAny3,
				},
				Operation: service.OperationType_AND,
			}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
		}

		_, _, err3 := mockDAO.Add(&idm.User{
			Login:     "admin",
			Password:  "xxxxxxx",
			GroupPath: "/path/to/group",
			Attributes: map[string]string{
				idm.UserAttrDisplayName: "Administrator",
				idm.UserAttrHidden:      "false",
				"active":                "true",
			},
			Roles: []*idm.Role{
				{Uuid: "1", Label: "Role1"},
				{Uuid: "4", Label: "Role4"},
			},
		})

		So(err3, ShouldBeNil)

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				HasRole: "1",
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 2)

			total, e2 := mockDAO.Count(&service.Query{SubQueries: []*anypb.Any{userQueryAny}})
			So(e2, ShouldBeNil)
			So(total, ShouldEqual, 2)

		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				HasRole: "1",
			}
			userQueryAny, _ := anypb.New(userQuery)

			userQuery2 := &idm.UserSingleQuery{
				HasRole: "2",
				Not:     true,
			}
			userQueryAny2, _ := anypb.New(userQuery2)

			e := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{
					userQueryAny,
					userQueryAny2,
				},
				Operation: service.OperationType_AND,
			}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 1)
			for _, user := range *users {
				So((user.(*idm.User)).Login, ShouldEqual, "admin")
				break
			}

		}

		{
			users := new([]interface{})
			userQueryAny, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/",
				Recursive: true,
			})
			e := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{userQueryAny},
			}, users)
			So(e, ShouldBeNil)
			So(users, ShouldHaveLength, 6)
			log.Print(users)
			allGroups := []*idm.User{}
			allUsers := []*idm.User{}
			for _, u := range *users {
				obj := u.(*idm.User)
				if obj.IsGroup {
					allGroups = append(allGroups, obj)
				} else {
					allUsers = append(allUsers, obj)
				}
			}
			So(allGroups, ShouldHaveLength, 4)
			So(allUsers, ShouldHaveLength, 2)
		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				Login: "username",
			}
			userQueryAny, _ := anypb.New(userQuery)
			mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			u := (*users)[0].(*idm.User)
			So(u, ShouldNotBeNil)
			// Change groupPath
			So(u.GroupPath, ShouldEqual, "/path/to/group/")
			// Move User
			u.GroupPath = "/path/to/anotherGroup"
			addedUser, _, e := mockDAO.Add(u)
			So(e, ShouldBeNil)
			So(addedUser.(*idm.User).GroupPath, ShouldEqual, "/path/to/anotherGroup")
			So(addedUser.(*idm.User).Login, ShouldEqual, "username")

			users2 := new([]interface{})
			userQueryAny2, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/path/to/anotherGroup",
			})
			e2 := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{userQueryAny2},
			}, users2)
			So(e2, ShouldBeNil)
			So(users2, ShouldHaveLength, 1)

			users3 := new([]interface{})
			userQueryAny3, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/path/to/group",
			})
			e3 := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{userQueryAny3},
			}, users3)
			So(e3, ShouldBeNil)
			So(users3, ShouldHaveLength, 1)

		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				FullPath: "/path/to/anotherGroup",
			}
			userQueryAny, _ := anypb.New(userQuery)
			mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			u := (*users)[0].(*idm.User)
			So(u, ShouldNotBeNil)
			// Change groupPath
			So(u.IsGroup, ShouldBeTrue)
			// Move Group
			u.GroupPath = "/anotherGroup"
			addedGroup, _, e := mockDAO.Add(u)
			So(e, ShouldBeNil)
			So(addedGroup.(*idm.User).GroupPath, ShouldEqual, "/anotherGroup")

			users2 := new([]interface{})
			userQueryAny2, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/path/to/anotherGroup",
			})
			e2 := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{userQueryAny2},
			}, users2)
			So(e2, ShouldBeNil)
			So(users2, ShouldHaveLength, 0)

			users3 := new([]interface{})
			userQueryAny3, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/anotherGroup",
			})
			e3 := mockDAO.Search(&service.Query{
				SubQueries: []*anypb.Any{userQueryAny3},
			}, users3)
			So(e3, ShouldBeNil)
			So(users3, ShouldHaveLength, 1)

		}

		{
			users := new([]interface{})
			userQuery := &idm.UserSingleQuery{
				GroupPath: "/",
				Recursive: false,
			}
			userQueryAny, _ := anypb.New(userQuery)

			e := mockDAO.Search(&service.Query{SubQueries: []*anypb.Any{userQueryAny}}, users)
			So(e, ShouldBeNil)
			for _, u := range *users {
				log.Print(u)
			}
			So(users, ShouldHaveLength, 2)
		}

		{
			// Delete a group
			userQueryAny3, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/anotherGroup",
			})
			num, e3 := mockDAO.Del(&service.Query{SubQueries: []*anypb.Any{userQueryAny3}}, make(chan *idm.User, 100))
			So(e3, ShouldBeNil)
			So(num, ShouldEqual, 2)
		}

		{
			// Delete all should be prevented
			_, e3 := mockDAO.Del(&service.Query{}, make(chan *idm.User, 100))
			So(e3, ShouldNotBeNil)
		}

		{
			// Delete a user
			userQueryAny3, _ := anypb.New(&idm.UserSingleQuery{
				GroupPath: "/path/to/group/",
				Login:     "admin",
			})
			num, e3 := mockDAO.Del(&service.Query{SubQueries: []*anypb.Any{userQueryAny3}}, make(chan *idm.User, 100))
			So(e3, ShouldBeNil)
			So(num, ShouldEqual, 1)
		}

	})

	Convey("Query Builder W/ subquery", t, func() {

		singleQ1, singleQ2, singleQ3 := new(idm.UserSingleQuery), new(idm.UserSingleQuery), new(idm.UserSingleQuery)

		singleQ1.Login = "user1"
		singleQ2.Login = "user2"
		singleQ3.Login = "user3"

		singleQ1Any, err := anypb.New(singleQ1)
		So(err, ShouldBeNil)

		singleQ2Any, err := anypb.New(singleQ2)
		So(err, ShouldBeNil)

		singleQ3Any, err := anypb.New(singleQ3)
		So(err, ShouldBeNil)

		subQuery1 := &service.Query{
			SubQueries: []*anypb.Any{singleQ1Any, singleQ2Any},
			Operation:  service.OperationType_OR,
		}

		subQuery2 := &service.Query{
			SubQueries: []*anypb.Any{singleQ3Any},
		}

		subQuery1Any, err := anypb.New(subQuery1)
		So(err, ShouldBeNil)
		test := subQuery1Any.MessageIs(new(service.Query))
		So(test, ShouldBeTrue)

		subQuery2Any, err := anypb.New(subQuery2)
		So(err, ShouldBeNil)

		composedQuery := &service.Query{
			SubQueries: []*anypb.Any{
				subQuery1Any,
				subQuery2Any,
			},
			Offset:    0,
			Limit:     10,
			Operation: service.OperationType_AND,
		}

		s := sql.NewQueryBuilder(composedQuery, converter).Expression("sqlite")
		So(s, ShouldNotBeNil)
		//So(s, ShouldEqual, "((t.uuid = n.uuid and (n.name='user1' and n.leaf = 1)) OR (t.uuid = n.uuid and (n.name='user2' and n.leaf = 1))) AND (t.uuid = n.uuid and (n.name='user3' and n.leaf = 1))")
	})
}

func TestDestructiveCreateUser(t *testing.T) {
	var options = configx.New()
	ctx := context.Background()
	var mock DAO
	if d, e := dao.InitDAO(ctx, sqlite.Driver, sqlite.SharedMemDSN, "idm_user", NewDAO, options); e != nil {
		panic(e)
	} else {
		mock = d.(DAO)
	}

	Convey("Test bug with create user", t, func() {

		_, _, err := mock.Add(&idm.User{
			Login:     "username",
			Password:  "xxxxxxx",
			GroupPath: "/path/to/group",
			Attributes: map[string]string{
				idm.UserAttrDisplayName: "John Doe",
				idm.UserAttrHidden:      "false",
				"active":                "true",
			},
			Roles: []*idm.Role{
				{Uuid: "1", Label: "Role1"},
				{Uuid: "2", Label: "Role2"},
			},
		})

		So(err, ShouldBeNil)

		_, _, err = mock.Add(&idm.User{
			Uuid:     "fixed-uuid",
			Login:    "",
			Password: "hashed",
		})

		So(err, ShouldNotBeNil)

		var target []interface{}
		ch := mock.GetNodeTree(context.Background(), mtree.NewMPath(1))
		for n := range ch {
			tn := n.(*mtree.TreeNode)
			t.Logf("Got node %s (%s)", tn.MPath.String(), tn.Name())
			target = append(target, n)
		}
		So(target, ShouldNotBeEmpty)

		_, _, err = mock.Add(&idm.User{
			Uuid:     "fixed-uuid",
			Login:    "",
			Password: "hashed",
		})

		//So(err, ShouldNotBeNil)

		var target2 []interface{}
		ch2 := mock.GetNodeTree(context.Background(), mtree.NewMPath(1))
		for n := range ch2 {
			tn := n.(*mtree.TreeNode)
			t.Logf("Got node %s (%s)", tn.MPath.String(), tn.Name())
			target2 = append(target2, n)
		}
		So(target2, ShouldNotBeEmpty)

	})

}
