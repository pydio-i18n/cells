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

package path

import (
	"context"
	"strings"
	"testing"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/config"
	"github.com/pydio/cells/v5/common/errors"
	"github.com/pydio/cells/v5/common/nodes"
	"github.com/pydio/cells/v5/common/nodes/models"
	"github.com/pydio/cells/v5/common/proto/idm"
	"github.com/pydio/cells/v5/common/proto/tree"
	"github.com/pydio/cells/v5/common/utils/cache"
	"github.com/pydio/cells/v5/common/utils/cache/gocache"
	cache_helper "github.com/pydio/cells/v5/common/utils/cache/helper"
	"github.com/pydio/cells/v5/common/utils/openurl"
	"github.com/pydio/cells/v5/common/utils/propagator"

	_ "github.com/pydio/cells/v5/common/config/memory"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx context.Context
)

func TestMain(m *testing.M) {
	cache_helper.SetStaticResolver("pm://?evictionTime=20s&cleanWindow=10s", &gocache.URLOpener{})
	mem, _ := config.OpenStore(context.Background(), "mem://")
	ctx = propagator.With(context.Background(), config.ContextKey, mem)
	nodes.SetSourcesPoolOpener(func(ctx context.Context) *openurl.Pool[nodes.SourcesPool] {
		return nodes.NewTestPool(ctx, nodes.MakeFakeClientsPool(nil, nil))
	})
	m.Run()
}

func newTestHandlerBranchTranslator() (*DataSourceHandler, *nodes.HandlerMock) {

	testRootNode := &tree.Node{
		Uuid:      "root-node-uuid",
		Path:      "datasource/root",
		MetaStore: make(map[string]string),
	}
	testRootNode.MustSetMeta(common.MetaNamespaceDatasourceName, "datasource")
	testRootNode.MustSetMeta(common.MetaNamespaceDatasourcePath, "root")
	b := newDataSourceHandler()
	ka := cache_helper.MustResolveCache(context.Background(), "any", cache.Config{})
	_ = ka.Set("root-node-uuid", testRootNode)
	mock := nodes.NewHandlerMock()
	mock.Nodes["datasource/root/inner/path"] = &tree.Node{
		Path: "datasource/root/inner/path",
		Uuid: "found-uuid",
	}
	mock.Nodes["datasource/root/inner/path/file"] = &tree.Node{
		Path: "datasource/root/inner/path/file",
		Uuid: "other-uuid",
	}
	b.SetNextHandler(mock)

	return b, mock

}

func makeFakeTestContext(identifier string, root ...*tree.Node) context.Context {

	fakeRoot := &tree.Node{Path: "datasource/root"}
	fakeRoot.MustSetMeta(common.MetaNamespaceDatasourceName, "datasource")
	b := nodes.BranchInfo{
		Workspace: &idm.Workspace{
			UUID:  "test-workspace",
			Slug:  "test-workspace",
			Label: "Test Workspace",
		},
		Root: fakeRoot,
	}
	if len(root) > 0 {
		b.Root = root[0]
	}
	return nodes.WithBranchInfo(ctx, identifier, b)

}

func TestBranchTranslator_ReadNode(t *testing.T) {

	Convey("Test Readnode without context", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		_, e := b.ReadNode(ctx, &tree.ReadNodeRequest{})
		So(errors.Is(e, errors.BranchInfoMissing), ShouldBeTrue)

	})

	Convey("Test Readnode with wrong context", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		c := nodes.WithBranchInfo(ctx, "in", nodes.BranchInfo{
			Workspace: &idm.Workspace{
				UUID:  "another-workspace",
				Label: "Another Workspace",
			},
		})
		_, e := b.ReadNode(c, &tree.ReadNodeRequest{})
		So(e, ShouldNotBeNil)
		So(errors.Is(e, errors.StatusInternalServerError), ShouldBeTrue)

	})

	Convey("Test Readnode with admin context", t, func() {

		b, mock := newTestHandlerBranchTranslator()
		adminCtx := nodes.WithBranchInfo(ctx, "in", nodes.BranchInfo{
			Workspace: &idm.Workspace{UUID: "ROOT"},
		})
		_, e := b.ReadNode(adminCtx, &tree.ReadNodeRequest{Node: &tree.Node{
			Path:      "datasource/root/path",
			MetaStore: make(map[string]string),
		}})
		So(e, ShouldNotBeNil)
		belowNode := mock.Nodes["in"]
		So(belowNode.Path, ShouldEqual, "datasource/root/path")
		So(belowNode.GetStringMeta(common.MetaNamespaceDatasourcePath), ShouldEqual, "root/path")
		outputBranch, er := nodes.GetBranchInfo(mock.Context, "in")
		So(er, ShouldBeNil)
		So(outputBranch.LoadedSource.ObjectsBucket, ShouldEqual, "bucket")

	})

	Convey("Test Readnode with user context", t, func() {

		b, mock := newTestHandlerBranchTranslator()

		ctx := makeFakeTestContext("in")
		resp, er := b.ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{
			Path:      "datasource/root/inner/path",
			MetaStore: make(map[string]string),
		}})
		So(er, ShouldBeNil) // Not found
		So(resp.Node.Path, ShouldEqual, "datasource/root/inner/path")
		So(resp.Node.Uuid, ShouldEqual, "found-uuid")

		belowNode := mock.Nodes["in"]
		So(belowNode, ShouldNotBeNil)
		So(belowNode.Path, ShouldEqual, "datasource/root/inner/path")
		//So(belowNode.GetStringMeta(common.MetaNamespaceDatasourcePath), ShouldEqual, "inner/path")
		outputBranch, er := nodes.GetBranchInfo(mock.Context, "in")
		So(er, ShouldBeNil)
		So(outputBranch.Workspace.UUID, ShouldEqual, "test-workspace")
		So(outputBranch.ObjectsBucket, ShouldEqual, "bucket")
	})

	Convey("Test update Output Node", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("in", &tree.Node{Path: "datasource/root"})
		node := &tree.Node{Path: "datasource/root/sub/path"}
		b.updateOutputNode(ctx, node, "in")
		So(node.Path, ShouldEqual, "datasource/root/sub/path")

	})

	Convey("Test update Output Node - Admin", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		adminCtx := nodes.WithBranchInfo(ctx, "in", nodes.BranchInfo{
			Workspace: &idm.Workspace{UUID: "ROOT"},
		})
		node := &tree.Node{Path: "datasource/root/sub/path"}
		b.updateOutputNode(adminCtx, node, "in")
		So(node.Path, ShouldEqual, "datasource/root/sub/path")

	})

}

func TestBranchTranslator_ListNodes(t *testing.T) {

	Convey("Test ListNodes with user context", t, func() {

		b, _ := newTestHandlerBranchTranslator()

		ctx := makeFakeTestContext("in")
		client, er := b.ListNodes(ctx, &tree.ListNodesRequest{Node: &tree.Node{
			Path:      "test-workspace/inner/path",
			MetaStore: make(map[string]string),
		}})
		So(er, ShouldBeNil) // found
		defer client.CloseSend()
		for {
			resp, e := client.Recv()
			if e != nil {
				break
			}
			if resp == nil {
				continue
			}
			So(resp.Node.Path, ShouldEqual, "test-workspace/inner/path/file")
			So(resp.Node.Uuid, ShouldEqual, "other-uuid")
			break // Test One N Only
		}

	})
}

func TestBranchTranslator_OtherMethods(t *testing.T) {

	Convey("Test CreateNode", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("in")
		_, er := b.CreateNode(ctx, &tree.CreateNodeRequest{Node: &tree.Node{
			Path:      "test-workspace/inner/path",
			MetaStore: make(map[string]string),
		}})
		So(er, ShouldBeNil) // found

	})

	Convey("Test DeleteNode", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("in")
		_, er := b.DeleteNode(ctx, &tree.DeleteNodeRequest{Node: &tree.Node{
			Path:      "test-workspace/inner/path",
			MetaStore: make(map[string]string),
		}})
		So(er, ShouldBeNil) // found

	})

	Convey("Test GetObject", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("in")
		_, er := b.GetObject(ctx, &tree.Node{
			Path:      "datasource/root/inner/path",
			MetaStore: make(map[string]string),
		}, &models.GetRequestData{})
		So(er, ShouldBeNil) // found

	})

	Convey("Test PutObject", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("in")
		_, er := b.PutObject(ctx, &tree.Node{
			Path:      "test-workspace/inner/path",
			MetaStore: make(map[string]string),
		}, strings.NewReader(""), &models.PutRequestData{})
		So(er, ShouldBeNil) // found

	})

	Convey("Test CopyObject", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("from")
		bI, _ := nodes.GetBranchInfo(ctx, "from")
		ctx = nodes.WithBranchInfo(ctx, "to", bI)
		_, er := b.CopyObject(ctx, &tree.Node{
			Path:      "test-workspace/inner/path",
			MetaStore: make(map[string]string),
		}, &tree.Node{
			Path:      "test-workspace/inner/path1",
			MetaStore: make(map[string]string),
		}, &models.CopyRequestData{})
		So(er, ShouldBeNil) // found

	})

	Convey("Test UpdateNode", t, func() {

		b, _ := newTestHandlerBranchTranslator()
		ctx := makeFakeTestContext("from")
		bI, _ := nodes.GetBranchInfo(ctx, "from")
		ctx = nodes.WithBranchInfo(ctx, "to", bI)
		_, er := b.UpdateNode(ctx, &tree.UpdateNodeRequest{
			From: &tree.Node{
				Path:      "test-workspace/inner/path",
				MetaStore: make(map[string]string),
			}, To: &tree.Node{
				Path:      "test-workspace/inner/path1",
				MetaStore: make(map[string]string),
			},
		})
		So(er, ShouldBeNil) // found

	})

}

func TestBranchTranslator_Multipart(t *testing.T) {

	/*
		Convey("Branch Translator Multipart Function NOT IMPLEMENTED", t, func() {

			b, _ := newTestHandlerBranchTranslator(NewTestPool(false))
			c := context.Background()
			_, e1 := b.MultipartCreate(c, &tree.N{}, &MultipartRequestData{})
			So(errors.Parse(e1.Error()).Code, ShouldEqual, 400)

			_, e1 = b.MultipartComplete(c, &tree.N{}, "uploadId", []minio.CompletePart{})
			So(errors.Parse(e1.Error()).Code, ShouldEqual, 400)

			e1 = b.MultipartAbort(c, &tree.N{}, "uploadId", &MultipartRequestData{})
			So(errors.Parse(e1.Error()).Code, ShouldEqual, 400)

			_, e1 = b.MultipartList(c, "", &MultipartRequestData{})
			So(errors.Parse(e1.Error()).Code, ShouldEqual, 400)

			_, e1 = b.MultipartListObjectParts(c, &tree.N{}, "uploadId", 0, 0)
			So(errors.Parse(e1.Error()).Code, ShouldEqual, 400)

		})
	*/

}
