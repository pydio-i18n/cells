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

package idmtest

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/server/stubs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/idm/role"
	srv "github.com/pydio/cells/v4/idm/role/grpc"
	"github.com/pydio/cells/v4/x/configx"
)

type RolesStreamer struct {
	stubs.ClientServerStreamerCore
	service *UsersService
}

// Send implements SERVER method
func (u *RolesStreamer) Send(response *idm.SearchRoleResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *RolesStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*idm.SearchRoleResponse).Role = resp.(*idm.SearchRoleResponse).Role
		return nil
	} else {
		return io.EOF
	}
}

func NewRolesService(roles ...*idm.Role) (*RolesService, error) {
	sqlDao2 := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "idm_roles")
	if sqlDao2 == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockRDAO := role.NewDAO(sqlDao2)
	var options = configx.New()
	if err := mockRDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise roles DAO, error: ", err)
	}

	serv := &RolesService{
		RoleServiceServer: srv.NewHandler(nil, mockRDAO.(role.DAO)),
	}
	ctx := servicecontext.WithDAO(context.Background(), mockRDAO)
	for _, r := range roles {
		_, er := serv.RoleServiceServer.CreateRole(ctx, &idm.CreateRoleRequest{Role: r})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type RolesService struct {
	idm.RoleServiceServer
	DAO dao.DAO
}

func (u *RolesService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &RolesStreamer{}
	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.RoleServiceServer.SearchRole(i.(*idm.SearchRoleRequest), st)
	}
	return st
}

func (u *RolesService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	var e error
	switch method {
	case "/idm.RoleService/CreateRole":
		if r, er := u.RoleServiceServer.CreateRole(ctx, args.(*idm.CreateRoleRequest)); er != nil {
			e = er
		} else {
			reply.(*idm.CreateRoleResponse).Role = r.GetRole()
		}
	default:
		return fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *RolesService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	switch method {
	case "/idm.RoleService/SearchRole":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf("not implemented")
}
