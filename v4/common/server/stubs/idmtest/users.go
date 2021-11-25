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

	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/server/stubs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/idm/user"
	srv "github.com/pydio/cells/v4/idm/user/grpc"
	"github.com/pydio/cells/v4/x/configx"
)

type UsersStreamer struct {
	stubs.ClientServerStreamerCore
	service *UsersService
}

// Send implements SERVER method
func (u *UsersStreamer) Send(response *idm.SearchUserResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *UsersStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*idm.SearchUserResponse).User = resp.(*idm.SearchUserResponse).User
		return nil
	} else {
		return io.EOF
	}
}

func NewUsersService(users ...*idm.User) (*UsersService, error) {
	sqlDao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "idm_user")
	if sqlDao == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockDAO := user.NewDAO(sqlDao)
	var options = configx.New()
	if err := mockDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise Users DAO, error: ", err)
	}

	serv := &UsersService{
		UserServiceServer: srv.NewHandler(nil, mockDAO.(user.DAO)),
	}
	ctx := servicecontext.WithDAO(context.Background(), mockDAO)
	for _, u := range users {
		_, er := serv.UserServiceServer.CreateUser(ctx, &idm.CreateUserRequest{User: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type UsersService struct {
	idm.UserServiceServer
}

func (u *UsersService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &UsersStreamer{}
	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.UserServiceServer.SearchUser(i.(*idm.SearchUserRequest), st)
	}
	return st
}

func (u *UsersService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	var e error
	switch method {
	case "/idm.UserService/CreateUser":
		if r, er := u.UserServiceServer.CreateUser(ctx, args.(*idm.CreateUserRequest)); er != nil {
			e = er
		} else {
			reply.(*idm.CreateUserResponse).User = r.GetUser()
		}
	case "/idm.UserService/DeleteUser":
		if r, er := u.UserServiceServer.DeleteUser(ctx, args.(*idm.DeleteUserRequest)); er != nil {
			e = er
		} else {
			reply.(*idm.DeleteUserResponse).RowsDeleted = r.GetRowsDeleted()
		}
	default:
		return fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *UsersService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	switch method {
	case "/idm.UserService/SearchUser":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf("not implemented")
}
