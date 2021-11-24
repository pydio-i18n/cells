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
	"github.com/pydio/cells/v4/idm/acl"
	srv "github.com/pydio/cells/v4/idm/acl/grpc"
	"github.com/pydio/cells/v4/x/configx"
)

type ACLStreamer struct {
	stubs.ClientServerStreamerCore
	service *ACLService
}

// Send implements SERVER method
func (u *ACLStreamer) Send(response *idm.SearchACLResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *ACLStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*idm.SearchACLResponse).ACL = resp.(*idm.SearchACLResponse).ACL
		return nil
	} else {
		return io.EOF
	}
}

func NewACLService(acls ...*idm.ACL) (*ACLService, error) {
	sqlDao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "idm_acl")
	if sqlDao == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockDAO := acl.NewDAO(sqlDao)
	var options = configx.New()
	if err := mockDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise DAO, error: ", err)
	}

	serv := &ACLService{
		DAO: mockDAO,
	}
	ctx := servicecontext.WithDAO(context.Background(), mockDAO)
	for _, u := range acls {
		_, er := serv.Handler.CreateACL(ctx, &idm.CreateACLRequest{ACL: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type ACLService struct {
	srv.Handler
	DAO dao.DAO
}

func (u *ACLService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &ACLStreamer{}
	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.Handler.SearchACL(i.(*idm.SearchACLRequest), st)
	}
	return st
}

func (u *ACLService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	var e error
	switch method {
	case "/idm.ACLService/CreateACL":
		if r, er := u.Handler.CreateACL(ctx, args.(*idm.CreateACLRequest)); er != nil {
			e = er
		} else {
			reply.(*idm.CreateACLResponse).ACL = r.GetACL()
		}
	}
	return e
}

func (u *ACLService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	switch method {
	case "/idm.ACLService/SearchACL":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf("not implemented")
}
