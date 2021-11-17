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

package grpc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/grpc/metadata"

	proto "github.com/pydio/cells/v4/common/proto/docstore"
	"github.com/pydio/cells/v4/data/docstore"
	. "github.com/smartystreets/goconvey/convey"
)

type listDocsTestStreamer struct {
	Docs []*proto.ListDocumentsResponse
	Ctx  context.Context
}

func (l *listDocsTestStreamer) SetHeader(md metadata.MD) error {
	panic("implement me")
}

func (l *listDocsTestStreamer) SendHeader(md metadata.MD) error {
	panic("implement me")
}

func (l *listDocsTestStreamer) SetTrailer(md metadata.MD) {
	panic("implement me")
}

func (l *listDocsTestStreamer) Context() context.Context {
	return l.Ctx
}

func newPath(tmpName string) string {
	return filepath.Join(os.TempDir(), tmpName)
}

func (l *listDocsTestStreamer) SendMsg(interface{}) error {
	return nil
}

func (l *listDocsTestStreamer) RecvMsg(interface{}) error {
	return nil
}

func (l *listDocsTestStreamer) Close() error {
	return nil
}

func (l *listDocsTestStreamer) Send(r *proto.ListDocumentsResponse) error {
	l.Docs = append(l.Docs, r)
	return nil
}

func createTestHandler(suffix string) *Handler {

	pBolt := newPath("docstore" + suffix + ".db")
	pBleve := newPath("docstore" + suffix + ".bleve")

	store, _ := docstore.NewBoltStore(pBolt, true)
	indexer, _ := docstore.NewBleveEngine(pBleve, true)

	return &Handler{
		Db:      store,
		Indexer: indexer,
	}

}

func TestHandler_Close(t *testing.T) {

	Convey("Test init/close handler", t, func() {

		pBolt := newPath("docstoreT.db")
		pBleve := newPath("docstoreT.bleve")
		defer func() {
			os.RemoveAll(pBolt)
			os.RemoveAll(pBleve)
		}()

		store, err := docstore.NewBoltStore(pBolt, true)
		So(err, ShouldBeNil)

		indexer, err := docstore.NewBleveEngine(pBleve, true)
		So(err, ShouldBeNil)

		handler := &Handler{
			Db:      store,
			Indexer: indexer,
		}

		cE := handler.Close()
		So(cE, ShouldBeNil)

		s1, _ := os.Stat(pBolt)
		s2, _ := os.Stat(pBleve)
		So(s1, ShouldBeNil)
		So(s2, ShouldBeNil)

	})

}

func TestHandler_CRUD(t *testing.T) {

	ctx := context.Background()
	Convey("Test Document GET/PUT/DELETE", t, func() {

		h := createTestHandler("crud")
		defer h.Close()

		_, e := h.PutDocument(ctx, &proto.PutDocumentRequest{
			StoreID: "any-store",
			Document: &proto.Document{
				Type:          proto.DocumentType_JSON,
				ID:            "my-doc-id",
				Owner:         "admin",
				Data:          "Hello World Data",
				IndexableMeta: `{"key1":"value1"}`,
			},
		})
		So(e, ShouldBeNil)

		getDocResp, e2 := h.GetDocument(ctx, &proto.GetDocumentRequest{
			StoreID:    "any-store",
			DocumentID: "my-doc-id",
		})
		So(e2, ShouldBeNil)
		So(getDocResp.Document.Data, ShouldResemble, "Hello World Data")

		delDocResp, e3 := h.DeleteDocuments(ctx, &proto.DeleteDocumentsRequest{
			StoreID:    "any-store",
			DocumentID: "my-doc-id",
		})
		So(e3, ShouldBeNil)
		So(delDocResp.Success, ShouldBeTrue)
		So(delDocResp.DeletionCount, ShouldEqual, 1)

		// Try to get deleted doc
		getDocResp2, e4 := h.GetDocument(ctx, &proto.GetDocumentRequest{
			StoreID:    "any-store",
			DocumentID: "my-doc-id",
		})
		So(e4, ShouldNotBeNil)
		So(getDocResp2, ShouldBeNil)

	})

}

func TestHandler_Search(t *testing.T) {

	ctx := context.Background()
	Convey("Test Document LIST/SEARCH", t, func() {

		h := createTestHandler("list")
		defer h.Close()

		_, e := h.PutDocument(ctx, &proto.PutDocumentRequest{
			StoreID: "any-store",
			Document: &proto.Document{
				Type:          proto.DocumentType_JSON,
				ID:            "my-doc-id-1",
				Owner:         "admin",
				Data:          "Hello World Data",
				IndexableMeta: `{"key":"value", "key2":"value2", "key3":45}`,
			},
		})
		So(e, ShouldBeNil)

		_, e = h.PutDocument(ctx, &proto.PutDocumentRequest{
			StoreID: "any-store",
			Document: &proto.Document{
				Type:          proto.DocumentType_JSON,
				ID:            "my-doc-id-2",
				Owner:         "charles",
				Data:          "Other Test Data",
				IndexableMeta: `{"key":"value", "key2":"other", "key3":50}`,
			},
		})
		So(e, ShouldBeNil)

		streamer := &listDocsTestStreamer{Ctx: ctx}
		e1 := h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				Owner: "admin",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 1)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				Owner: "unknwown",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 0)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key:value",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 2)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key2:value2",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 1)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key3:<49",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 1)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key3:<45",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 0)

		streamer = &listDocsTestStreamer{Ctx: ctx}
		e1 = h.ListDocuments(&proto.ListDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key3:>45 +key2:value2",
			},
		}, streamer)
		So(e1, ShouldBeNil)
		So(streamer.Docs, ShouldHaveLength, 0)

		// NOW DELETE DOCS
		delDocs, e1 := h.DeleteDocuments(ctx, &proto.DeleteDocumentsRequest{
			StoreID: "any-store",
			Query: &proto.DocumentQuery{
				MetaQuery: "+key:value",
			},
		})
		So(e1, ShouldBeNil)
		So(delDocs.DeletionCount, ShouldEqual, 2)

	})
}
