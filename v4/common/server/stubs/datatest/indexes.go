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

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/object"
	"github.com/pydio/cells/v4/data/source/index"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	_ "gopkg.in/doug-martin/goqu.v4/adapters/sqlite3"

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/server/stubs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/sql"
	srv "github.com/pydio/cells/v4/data/source/index/grpc"
	"github.com/pydio/cells/v4/x/configx"
)

type IndexStreamer struct {
	stubs.ClientServerStreamerCore
	service *IndexService
}

// Send implements SERVER method
func (u *IndexStreamer) Send(response *tree.ListNodesResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *IndexStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*tree.ListNodesResponse).Node = resp.(*tree.ListNodesResponse).Node
		return nil
	} else {
		return io.EOF
	}
}

func NewIndexService(dsName string, nodes ...*tree.Node) (*IndexService, error) {
	sqlDao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "data_index_"+dsName)
	if sqlDao == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockDAO := index.NewDAO(sqlDao)
	var options = configx.New()
	if err := mockDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise index DAO, error: ", err)
	}

	ts := srv.NewTreeServer(&object.DataSource{Name: dsName}, mockDAO.(index.DAO), log.Logger(context.Background()))

	serv := &IndexService{
		TreeServer: *ts,
		DAO:        mockDAO,
	}
	ctx := servicecontext.WithDAO(context.Background(), mockDAO)
	for _, u := range nodes {
		_, er := serv.TreeServer.CreateNode(ctx, &tree.CreateNodeRequest{Node: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type IndexService struct {
	srv.TreeServer
	DAO dao.DAO
}

func (u *IndexService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &IndexStreamer{}

	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.TreeServer.ListNodes(i.(*tree.ListNodesRequest), st)
	}
	return st
}

func (u *IndexService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
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

func (u *IndexService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	switch method {
	case "/tree.NodeProvider/ListNodes":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf(method + "  not implemented")
}
