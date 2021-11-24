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

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/server/stubs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/pydio/cells/v4/data/meta"
	srv "github.com/pydio/cells/v4/data/meta/grpc"
	"github.com/pydio/cells/v4/x/configx"
	"google.golang.org/grpc"
)

type ListMetaStreamer struct {
	stubs.ClientServerStreamerCore
	service *MetaService
}

// Send implements SERVER method
func (u *ListMetaStreamer) Send(response *tree.ListNodesResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *ListMetaStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*tree.ListNodesResponse).Node = resp.(*tree.ListNodesResponse).Node
		return nil
	} else {
		return io.EOF
	}
}

type ReadStreamMetaStreamer struct {
	stubs.ClientServerStreamerCore
	service *MetaService
	ReqChan chan interface{}
}

// Recv implements SERVER method
func (u *ReadStreamMetaStreamer) Recv() (*tree.ReadNodeRequest, error) {
	if req, o := <-u.ReqChan; o {
		return req.(*tree.ReadNodeRequest), nil
	} else {
		return nil, io.EOF
	}
}

// Send implements SERVER method
func (u *ReadStreamMetaStreamer) Send(response *tree.ReadNodeResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *ReadStreamMetaStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*tree.ReadNodeResponse).Node = resp.(*tree.ReadNodeResponse).Node
		return nil
	} else {
		return io.EOF
	}
}

func NewMetaService(nodes ...*tree.Node) (*MetaService, error) {
	sqlDao := sql.NewDAO("sqlite3", "file::memory:?mode=memory&cache=shared", "data_meta_")
	if sqlDao == nil {
		return nil, fmt.Errorf("unable to open sqlite3 DB file, could not start test")
	}

	mockDAO := meta.NewDAO(sqlDao)
	var options = configx.New()
	if err := mockDAO.Init(options); err != nil {
		return nil, fmt.Errorf("could not start test: unable to initialise index DAO, error: ", err)
	}

	ts := srv.NewMetaServer(context.Background())

	serv := &MetaService{
		MetaServer: *ts,
		DAO:        mockDAO,
	}
	ctx := servicecontext.WithDAO(context.Background(), mockDAO)
	for _, u := range nodes {
		_, er := serv.MetaServer.CreateNode(ctx, &tree.CreateNodeRequest{Node: u})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type MetaService struct {
	srv.MetaServer
	DAO dao.DAO
}

func (u *MetaService) GetStreamer(ctx context.Context, streamType string) grpc.ClientStream {

	if streamType == "list" {
		st := &ListMetaStreamer{}
		st.Ctx = ctx
		st.RespChan = make(chan interface{}, 1000)
		st.SendHandler = func(i interface{}) error {
			return u.MetaServer.ListNodes(i.(*tree.ListNodesRequest), st)
		}
		return st
	} else {
		st := &ReadStreamMetaStreamer{}
		st.Ctx = ctx
		st.RespChan = make(chan interface{}, 1000)
		st.ReqChan = make(chan interface{}, 1000)
		st.SendHandler = func(i interface{}) error {
			//return u.MetaServer.ReadNodeStream(st)
			st.ReqChan <- i
			return nil
		}
		go u.MetaServer.ReadNodeStream(st)
		return st
	}
}

func (u *MetaService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	var e error
	switch method {
	case "/tree.NodeReceiver/CreateNode":
		resp, er := u.MetaServer.CreateNode(ctx, args.(*tree.CreateNodeRequest))
		if er == nil {
			reply.(*tree.CreateNodeResponse).Node = resp.GetNode()
			reply.(*tree.CreateNodeResponse).Success = resp.GetSuccess()
		} else {
			e = er
		}
	default:
		e = fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *MetaService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = servicecontext.WithDAO(ctx, u.DAO)
	switch method {
	case "/tree.NodeProvider/ListNodes":
		return u.GetStreamer(ctx, "list"), nil
	case "/tree.NodeProviderStreamer/ReadNodeStream":
		return u.GetStreamer(ctx, "stream"), nil
	}
	return nil, fmt.Errorf(method + "  not implemented")
}
