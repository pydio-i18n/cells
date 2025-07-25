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

package commons

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/auth/claim"
	"github.com/pydio/cells/v5/common/nodes"
	"github.com/pydio/cells/v5/common/nodes/compose"
	"github.com/pydio/cells/v5/common/nodes/meta"
	"github.com/pydio/cells/v5/common/proto/tree"
	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/common/storage/indexer"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/common/utils/configx"
	"github.com/pydio/cells/v5/common/utils/propagator"
	"github.com/pydio/cells/v5/data/search"
	"github.com/pydio/cells/v5/data/search/analyzers"
)

// LocalBatch avoids overflowing bleve index by batching indexation events (index/delete)
type LocalBatch struct {
	sync.Mutex
	inserts    map[string]*tree.IndexableNode
	deletes    map[string]struct{}
	nsProvider *meta.NsProvider
	options    BatchOptions
	ctx        context.Context
	uuidRouter nodes.Handler
	stdRouter  nodes.Handler
}

type BatchOptions struct {
	config configx.Values
}

func NewBatch(ctx context.Context, nsProvider *meta.NsProvider, options BatchOptions, inputOpts ...indexer.BatchOption) indexer.Batch {
	wrapper := &LocalBatch{
		options:    options,
		inserts:    make(map[string]*tree.IndexableNode),
		deletes:    make(map[string]struct{}),
		nsProvider: nsProvider,
	}
	wrapper.ctx = wrapper.createBackgroundContext(ctx)

	iOpts := append(inputOpts,
		indexer.WithFlushCondition(func() bool {
			return len(wrapper.inserts)+len(wrapper.deletes) > BatchSize
		}),
		indexer.WithInsertCallback(func(msg any) error {
			i, ok := msg.(*tree.IndexableNode)
			if !ok {
				return errors.New("wrong message in batch insert")
			}

			wrapper.Lock()
			wrapper.inserts[i.GetUuid()] = i
			delete(wrapper.deletes, i.GetUuid())
			wrapper.Unlock()
			return nil
		}),
		indexer.WithDeleteCallback(func(msg any) error {
			uuid, ok := msg.(string)
			if !ok {
				return errors.New("wrong message in batch delete")
			}

			wrapper.Lock()
			wrapper.deletes[uuid] = struct{}{}
			delete(wrapper.inserts, uuid)
			wrapper.Unlock()
			return nil
		}),
		indexer.WithFlushCallback(func() error {
			return wrapper.Flush(ctx, inputOpts...)
		}),
	)
	return indexer.NewBatch(ctx, iOpts...)
}

func (b *LocalBatch) Flush(ctx context.Context, batchOpts ...indexer.BatchOption) error {

	idx, er := manager.Resolve[search.Engine](ctx)
	if er != nil {
		return er
	}
	b.Lock()
	defer b.Unlock()
	l := len(b.inserts) + len(b.deletes)
	if l == 0 {
		return nil
	}
	log.Logger(b.ctx).Debug("Flushing search batch", zap.Int("size", l))
	excludes := b.nsProvider.ExcludeIndexes()
	var nn []*tree.IndexableNode
	if er := b.nsProvider.InitStreamers(b.ctx); er != nil {
		return er
	}
	for uuid, node := range b.inserts {
		if e := b.LoadIndexableNode(node, excludes); e == nil {
			nn = append(nn, node)
		}
		delete(b.inserts, uuid)
	}
	if er := b.nsProvider.CloseStreamers(); er != nil {
		return er
	}

	// Now create an indexer batch, fill it and directly flush it
	batch, err := idx.NewBatch(ctx, batchOpts...)
	if err != nil {
		return err
	}
	for _, n := range nn {
		if er = batch.Insert(n); er != nil {
			log.Logger(ctx).Warn("Search batch - InsertOne error", zap.Error(er))
		}
	}
	for uuid := range b.deletes {
		if er = batch.Delete(uuid); er != nil {
			log.Logger(ctx).Warn("Search batch - DeleteOne error", zap.Error(er))
		}
		delete(b.deletes, uuid)
	}
	if er = batch.Flush(); er != nil {
		log.Logger(ctx).Warn("Error while flushing local batch", zap.Error(er))
	}
	return batch.Close()

}

func (b *LocalBatch) LoadIndexableNode(indexNode *tree.IndexableNode, excludes map[string]struct{}) error {
	if indexNode.ReloadCore {
		if resp, e := b.getUuidRouter().ReadNode(b.ctx, &tree.ReadNodeRequest{Node: indexNode.Node}); e != nil {
			return e
		} else {
			rNode := resp.Node
			if indexNode.MetaStore != nil {
				for k, v := range indexNode.MetaStore {
					rNode.MetaStore[k] = v
				}
			}
			indexNode.Node = resp.GetNode()
		}
	} else if indexNode.ReloadNs {
		if resp, e := b.nsProvider.ReadNode(indexNode.Node); e != nil {
			return e
		} else {
			indexNode.Node = resp
		}
	}
	indexNode.PathDepth = len(strings.Split(strings.Trim(indexNode.Path, "/"), "/"))
	indexNode.Meta = indexNode.AllMetaDeserialized(excludes)
	indexNode.ModifTime = time.Unix(indexNode.MTime, 0)
	var basename string
	_ = indexNode.GetMeta(common.MetaNamespaceNodeName, &basename)
	indexNode.Basename = basename
	if indexNode.Type == 1 {
		indexNode.NodeType = "file"
		indexNode.Extension = strings.ToLower(strings.TrimLeft(filepath.Ext(basename), "."))
	} else {
		indexNode.NodeType = "folder"
	}
	// Apply custom analyzers
	if er := analyzers.Parse(b.ctx, indexNode, b.options.config); er != nil {
		return er
	}

	indexNode.MetaStore = nil
	return nil
}

func (b *LocalBatch) createBackgroundContext(parent context.Context) context.Context {
	ctx := claim.ToContext(context.Background(), claim.Claims{
		Name:      common.PydioSystemUsername,
		Profile:   common.PydioProfileAdmin,
		GroupPath: "/",
	})
	return propagator.ForkContext(ctx, parent)
}

func (b *LocalBatch) getUuidRouter() nodes.Handler {
	if b.uuidRouter == nil {
		b.uuidRouter = compose.UuidClient(nodes.AsAdmin())
	}
	return b.uuidRouter
}

func (b *LocalBatch) getStdRouter() nodes.Handler {
	if b.stdRouter == nil {
		b.stdRouter = compose.PathClientAdmin()
	}
	return b.stdRouter
}
