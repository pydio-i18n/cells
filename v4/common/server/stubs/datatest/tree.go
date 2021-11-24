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

package datatest

import (
	"context"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	_ "gopkg.in/doug-martin/goqu.v4/adapters/sqlite3"

	"github.com/pydio/cells/v4/common"
	grpc2 "github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/server/stubs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	srv "github.com/pydio/cells/v4/data/tree/grpc"
)

type TreeStreamer struct {
	stubs.ClientServerStreamerCore
	service *TreeService
}

// Send implements SERVER method
func (u *TreeStreamer) Send(response *tree.ListNodesResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *TreeStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*tree.ListNodesResponse).Node = resp.(*tree.ListNodesResponse).Node
		return nil
	} else {
		return io.EOF
	}
}

func NewTreeService(dss []string, nodes ...*tree.Node) (*TreeService, error) {

	serv := &TreeService{}
	serv.DataSources = map[string]srv.DataSource{}
	for _, ds := range dss {
		conn := grpc2.NewClientConn(common.ServiceDataIndex_ + ds)
		serv.DataSources[ds] = srv.NewDataSource(ds, tree.NewNodeProviderClient(conn), tree.NewNodeReceiverClient(conn))
	}

	for _, u := range nodes {
		_, er := serv.TreeServer.CreateNode(context.Background(), &tree.CreateNodeRequest{Node: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type TreeService struct {
	srv.TreeServer
	DAO dao.DAO
}

func (u *TreeService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &TreeStreamer{}

	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.TreeServer.ListNodes(i.(*tree.ListNodesRequest), st)
	}
	return st
}

func (u *TreeService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	var e error
	switch method {
	case "/tree.NodeReceiver/CreateNode":
		resp, er := u.TreeServer.CreateNode(ctx, args.(*tree.CreateNodeRequest))
		if er == nil {
			reply.(*tree.CreateNodeResponse).Node = resp.GetNode()
			reply.(*tree.CreateNodeResponse).Success = resp.GetSuccess()
		} else {
			e = er
		}
	case "/tree.NodeProvider/ReadNode":
		resp, er := u.TreeServer.ReadNode(ctx, args.(*tree.ReadNodeRequest))
		if er == nil {
			reply.(*tree.ReadNodeResponse).Node = resp.GetNode()
			reply.(*tree.ReadNodeResponse).Success = resp.GetSuccess()
		} else {
			e = er
		}
	default:
		e = fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *TreeService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	switch method {
	case "/tree.NodeProvider/ListNodes":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf(method + "  not implemented")
}
