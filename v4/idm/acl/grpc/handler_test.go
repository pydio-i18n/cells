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

package grpc

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	. "github.com/smartystreets/goconvey/convey"

	// SQLite Driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/pydio/cells/v4/common/proto/idm"
	service "github.com/pydio/cells/v4/common/proto/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/idm/acl"
	"github.com/pydio/cells/v4/x/configx"
)

var (
	ctx     context.Context
	mockDAO acl.DAO
	options = configx.New()

	wg sync.WaitGroup
)

func TestMain(m *testing.M) {

	dao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "test_")
	if dao == nil {
		fmt.Print("Could not start test")
		return
	}

	mockDAO = acl.NewDAO(dao).(acl.DAO)
	if err := mockDAO.Init(options); err != nil {
		fmt.Print("Could not start test ", err)
		return
	}

	ctx = servicecontext.WithDAO(context.Background(), mockDAO)

	m.Run()
	wg.Wait()
}

func TestACL(t *testing.T) {

	s := new(Handler)

	Convey("Create ACLs", t, func() {
		resp := new(idm.CreateACLResponse)
		err := s.CreateACL(ctx, &idm.CreateACLRequest{ACL: &idm.ACL{
			NodeID:      "fake-node-id",
			WorkspaceID: "fake-ws-id",
			Action:      &idm.ACLAction{Name: "read", Value: "1"},
			RoleID:      "role1"},
		}, resp)

		So(err, ShouldBeNil)
		So(resp.GetACL().GetID(), ShouldEqual, "1")

	})

	Convey("Create ACLs", t, func() {
		resp := new(idm.CreateACLResponse)
		err := s.CreateACL(ctx, &idm.CreateACLRequest{ACL: &idm.ACL{
			NodeID:      "fake-node-id",
			WorkspaceID: "fake-ws-id",
			Action:      &idm.ACLAction{Name: "read", Value: "1"},
			RoleID:      "role2"},
		}, resp)

		So(err, ShouldBeNil)
		So(resp.GetACL().GetID(), ShouldEqual, "2")
	})

	Convey("Get ACL", t, func() {
		mock := &aclStreamMock{}
		err := s.StreamACL(ctx, mock)

		So(err, ShouldBeNil)
		So(len(mock.InternalBuffer), ShouldEqual, 0)
	})

	Convey("Search ACL", t, func() {
		mock := &aclStreamMock{}
		readQ, _ := anypb.New(&idm.ACLSingleQuery{
			Actions: []*idm.ACLAction{{Name: "read"}},
		})
		err := s.SearchACL(ctx, &idm.SearchACLRequest{Query: &service.Query{SubQueries: []*anypb.Any{readQ}}}, mock)

		So(err, ShouldBeNil)
		So(len(mock.InternalBuffer), ShouldEqual, 2)
	})

	Convey("Del ACL", t, func() {
		err := s.DeleteACL(ctx, &idm.DeleteACLRequest{}, &idm.DeleteACLResponse{})

		So(err, ShouldNotBeNil)
	})

	Convey("Del ACL", t, func() {
		err := s.DeleteACL(ctx, &idm.DeleteACLRequest{}, &idm.DeleteACLResponse{})

		So(err, ShouldNotBeNil)
	})

	Convey("Del ACL", t, func() {
		singleQ1 := new(idm.ACLSingleQuery)
		singleQ1.RoleIDs = []string{"role1"}
		singleQ1Any, err := anypb.New(singleQ1)
		So(err, ShouldBeNil)

		query := &service.Query{
			SubQueries: []*anypb.Any{singleQ1Any},
		}

		err = s.DeleteACL(ctx, &idm.DeleteACLRequest{Query: query}, &idm.DeleteACLResponse{})
		So(err, ShouldBeNil)
	})

	Convey("Search ACL", t, func() {
		mock := &aclStreamMock{}
		err := s.SearchACL(ctx, &idm.SearchACLRequest{}, mock)

		So(err, ShouldBeNil)
		So(len(mock.InternalBuffer), ShouldEqual, 1)
	})
}

// =================================================
// * Mock *
// =================================================

type aclStreamMock struct {
	InternalBuffer []*idm.ACL
}

func (x *aclStreamMock) Context() context.Context {
	panic("implement me")
}

func (x *aclStreamMock) Close() error {
	return nil
}

func (x *aclStreamMock) SendMsg(m interface{}) error {
	return nil
}

func (x *aclStreamMock) RecvMsg(m interface{}) error {
	return nil
}

func (x *aclStreamMock) Recv() (*idm.SearchACLRequest, error) {
	return nil, nil
}

func (x *aclStreamMock) Send(m *idm.SearchACLResponse) error {
	x.InternalBuffer = append(x.InternalBuffer, m.ACL)
	return nil
}
