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
	"os"
	"path/filepath"

	"github.com/pydio/cells/v4/common"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/proto/docstore"
	"github.com/pydio/cells/v4/common/server/stubs"
	"github.com/pydio/cells/v4/common/utils/uuid"
	docstore2 "github.com/pydio/cells/v4/data/docstore"
	srv "github.com/pydio/cells/v4/data/docstore/grpc"
)

type DocStoreStreamer struct {
	stubs.ClientServerStreamerCore
	service *DocStoreService
}

// Send implements SERVER method
func (u *DocStoreStreamer) Send(response *docstore.ListDocumentsResponse) error {
	u.RespChan <- response
	return nil
}

// RecvMsg implements CLIENT method
func (u *DocStoreStreamer) RecvMsg(m interface{}) error {
	if resp, o := <-u.RespChan; o {
		m.(*docstore.ListDocumentsResponse).Document = resp.(*docstore.ListDocumentsResponse).Document
		return nil
	} else {
		return io.EOF
	}
}

func newPath(tmpName string) string {
	return filepath.Join(os.TempDir(), tmpName)
}

func defaults() map[string]string {

	return map[string]string{
		"my-files": `{"Uuid":"my-files","Path":"my-files","Type":"COLLECTION","MetaStore":{"name":"my-files", "onDelete":"rename-uuid","resolution":"\/\/ Default node used for storing personal users data in separate folders. \n\/\/ Use Ctrl+Space to see the objects available for completion.\nPath = DataSources.personal + \"\/\" + User.Name;","contentType":"text\/javascript"}}`,
		"cells":    `{"Uuid":"cells","Path":"cells","Type":"COLLECTION","MetaStore":{"name":"cells","resolution":"\/\/ Default node used as parent for creating empty cells. \n\/\/ Use Ctrl+Space to see the objects available for completion.\nPath = DataSources.cellsdata + \"\/\" + User.Name;","contentType":"text\/javascript"}}`,
	}

}

func NewDocStoreService() (*DocStoreService, error) {

	suffix := uuid.New()
	pBolt := newPath("docstore" + suffix + ".db")
	pBleve := newPath("docstore" + suffix + ".bleve")

	store, _ := docstore2.NewBoltStore(pBolt, true)
	indexer, _ := docstore2.NewBleveEngine(pBleve, true)

	h := &srv.Handler{
		Db:      store,
		Indexer: indexer,
	}

	serv := &DocStoreService{
		Handler: *h,
	}
	for id, json := range defaults() {
		_, er := serv.Handler.PutDocument(context.Background(), &docstore.PutDocumentRequest{
			StoreID:    common.DocStoreIdVirtualNodes,
			DocumentID: id,
			Document: &docstore.Document{
				ID:    id,
				Type:  docstore.DocumentType_JSON,
				Owner: common.PydioSystemUsername,
				Data:  json,
			},
		})
		if er != nil {
			return nil, er
		}
	}
	return serv, nil
}

type DocStoreService struct {
	srv.Handler
}

func (u *DocStoreService) GetStreamer(ctx context.Context) grpc.ClientStream {
	st := &DocStoreStreamer{}
	st.Ctx = ctx
	st.RespChan = make(chan interface{}, 1000)
	st.SendHandler = func(i interface{}) error {
		return u.Handler.ListDocuments(i.(*docstore.ListDocumentsRequest), st)
	}
	return st
}

func (u *DocStoreService) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	fmt.Println("Serving", method, args, reply, opts)
	var e error
	switch method {
	case "/docstore.DocStore/PutDocument":
		if r, er := u.Handler.PutDocument(ctx, args.(*docstore.PutDocumentRequest)); er != nil {
			e = er
		} else {
			reply.(*docstore.PutDocumentResponse).Document = r.GetDocument()
		}
	case "/docstore.DocStore/DeleteDocuments":
		if r, er := u.Handler.DeleteDocuments(ctx, args.(*docstore.DeleteDocumentsRequest)); er != nil {
			e = er
		} else {
			reply.(*docstore.DeleteDocumentsResponse).DeletionCount = r.DeletionCount
			reply.(*docstore.DeleteDocumentsResponse).Success = r.Success
		}
	default:
		e = fmt.Errorf(method + " not implemented")
	}
	return e
}

func (u *DocStoreService) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	switch method {
	case "/docstore.DocStore/ListDocuments":
		return u.GetStreamer(ctx), nil
	}
	return nil, fmt.Errorf(method + " not implemented")
}
