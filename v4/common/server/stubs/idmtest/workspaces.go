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
	"github.com/pydio/cells/v4/idm/workspace"
	srv "github.com/pydio/cells/v4/idm/workspace/grpc"
	"github.com/pydio/cells/v4/x/configx"
)

type WorkspacesStreamer struct {
	stubs.ClientServerStreamerCore
	service *WorkspacesService
}

// Send implements SERVER method
func (u *WorkspacesStreamer) Send(response *idm.SearchWorkspaceResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *WorkspacesStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*idm.SearchWorkspaceResponse).Workspace = resp.(*idm.SearchWorkspaceResponse).Workspace
		return nil
	} else {
		return io.EOF
	}
}

func NewWorkspacesService(ww ...*idm.Workspace) (*WorkspacesService, error) {
	sqlDao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "idm_workspace")
	if sqlDao == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockDAO := workspace.NewDAO(sqlDao)
	var options = configx.New()
	if err := mockDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise WS DAO, error: ", err)
	}

	serv := &WorkspacesService{
		WorkspaceServiceServer: srv.NewHandler(nil, mockDAO.(workspace.DAO)),
	}
	ctx := servicecontext.WithDAO(context.Background(), mockDAO)
	for _, u := range ww {
		_, er := serv.WorkspaceServiceServer.CreateWorkspace(ctx, &idm.CreateWorkspaceRequest{Workspace: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type WorkspacesService struct {
	idm.WorkspaceServiceServer
}

func (u *WorkspacesService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &WorkspacesStreamer{}
	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.WorkspaceServiceServer.SearchWorkspace(i.(*idm.SearchWorkspaceRequest), st)
	}
	return st
}

func (u *WorkspacesService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	var e error
	switch method {
	case "/idm.WorkspaceService/CreateWorkspace":
		cr, er := u.WorkspaceServiceServer.CreateWorkspace(ctx, args.(*idm.CreateWorkspaceRequest))
		if er != nil {
			e = er
		} else {
			reply.(*idm.CreateWorkspaceResponse).Workspace = cr.GetWorkspace()
		}
	case "/idm.WorkspaceService/DeleteWorkspace":
		cr, er := u.WorkspaceServiceServer.DeleteWorkspace(ctx, args.(*idm.DeleteWorkspaceRequest))
		if er != nil {
			e = er
		} else {
			reply.(*idm.DeleteWorkspaceResponse).RowsDeleted = cr.GetRowsDeleted()
		}
	default:
		return fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *WorkspacesService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	switch method {
	case "/idm.WorkspaceService/SearchWorkspace":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf("not implemented")
}
