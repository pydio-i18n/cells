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
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/nodes/meta"
	"github.com/pydio/cells/v4/common/nodes/mocks"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/runtime"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/service/errors"
)

// changesListener is an autoclosing pipe used for fanning out events
type changesListener struct {
	in      chan *tree.NodeChangeEvent
	out     chan *tree.NodeChangeEvent
	done    chan struct{}
	closing bool
}

func newListener() *changesListener {
	l := &changesListener{
		in:   make(chan *tree.NodeChangeEvent),
		out:  make(chan *tree.NodeChangeEvent, 1000),
		done: make(chan struct{}, 1),
	}
	go func() {
		defer close(l.in)
		defer close(l.out)
		for {
			select {
			case i := <-l.in:
				l.out <- i
			case <-l.done:
				return
			}
		}
	}()
	return l
}

func (l *changesListener) stop() {
	l.closing = true
	close(l.done)
}

type DataSource struct {
	Name   string
	writer tree.NodeReceiverClient
	reader tree.NodeProviderClient
}

func NewDataSource(name string, reader tree.NodeProviderClient, writer tree.NodeReceiverClient) DataSource {
	return DataSource{
		Name:   name,
		reader: reader,
		writer: writer,
	}
}

type TreeServer struct {
	tree.UnimplementedSearcherServer
	tree.UnimplementedNodeReceiverServer
	tree.UnimplementedNodeProviderServer
	tree.UnimplementedNodeProviderStreamerServer
	tree.UnimplementedNodeChangesStreamerServer

	name      string
	listeners []*changesListener

	sources     map[string]DataSource
	sourcesLock *sync.RWMutex
	mainCtx     context.Context
}

// NewTreeServer initialize a TreeServer with proper internals
func NewTreeServer(ctx context.Context, name string) *TreeServer {
	return &TreeServer{
		mainCtx:     ctx,
		name:        name,
		sources:     make(map[string]DataSource),
		sourcesLock: &sync.RWMutex{},
	}
}

func (s *TreeServer) Name() string {
	return s.name
}

// AppendDatasource feeds internal datasources map
func (s *TreeServer) AppendDatasource(name string, obj DataSource) {
	s.sourcesLock.Lock()
	s.sources[name] = obj
	s.sourcesLock.Unlock()
}

// datasourcebyName finds a datasource in the internal map
func (s *TreeServer) datasourceByName(dsName string) (DataSource, bool) {
	s.sourcesLock.RLock()
	ds, ok := s.sources[dsName]
	s.sourcesLock.RUnlock()
	return ds, ok
}

// ReadNodeStream Implement stream for readNode method
func (s *TreeServer) ReadNodeStream(streamer tree.NodeProviderStreamer_ReadNodeStreamServer) error {

	// In some cases, initial ctx could be canceled _before_ this function is called
	// We must make sure that metaStreamers are using a proper context at creation
	// otherwise it can create a goroutine leak on linux.
	ctx := metadata.NewBackgroundWithMetaCopy(streamer.Context())
	ctx = runtime.ForkContext(ctx, s.mainCtx)

	var flags tree.Flags
	if sf, o := metadata.CanonicalMeta(streamer.Context(), tree.StatFlagHeaderName); o {
		flags = tree.StatFlagsFromString(sf)
	}

	metaStreamer := meta.NewStreamLoader(ctx, flags)
	defer metaStreamer.Close()

	msCtx := context.WithValue(ctx, "MetaStreamer", metaStreamer)
	msCtx = context.WithValue(msCtx, "ServicesListReloaded", true)
	for {
		request, err := streamer.Recv()
		if request == nil {
			break
		}
		if err != nil {
			return err
		}
		response, err := s.ReadNode(msCtx, request)
		if err == nil {
			response.Success = true
		} else {
			response = &tree.ReadNodeResponse{}
		}
		if e := streamer.Send(response); e != nil {
			return e
		}
	}

	return nil
}

func (s *TreeServer) treeNodeToDataSourcePath(node *tree.Node) (dataSourceName string, dataSourcePath string) {

	path := strings.Trim(node.GetPath(), "/")
	if path == "" {
		return "", ""
	}
	parts := strings.Split(path, "/")
	dataSourceName = parts[0]
	if len(parts) > 1 {
		dataSourcePath = strings.Join(parts[1:], "/")
	} else {
		dataSourcePath = ""
	}

	return dataSourceName, dataSourcePath

}

func (s *TreeServer) updateDataSourceNode(node *tree.Node, dataSourceName string) {

	dsPath := strings.TrimLeft(node.GetPath(), "/")
	newPath := dataSourceName + "/" + dsPath

	node.Path = newPath
	node.MustSetMeta(common.MetaNamespaceDatasourcePath, dsPath)
	if node.Uuid == "ROOT" {
		node.Uuid = "DATASOURCE:" + dataSourceName
	}
}

/* =============================================================================
 *  Server public Methods
 * ============================================================================ */

// CreateNode implementation for the TreeServer
func (s *TreeServer) CreateNode(ctx context.Context, req *tree.CreateNodeRequest) (*tree.CreateNodeResponse, error) {
	log.Logger(ctx).Debug("Create Node", zap.String("UUID", req.Node.Uuid), zap.String("Path", req.Node.Path))
	node := req.GetNode()
	resp := &tree.CreateNodeResponse{}

	defer track("CreateNode", ctx, time.Now(), req, resp)

	dsName, dsPath := s.treeNodeToDataSourcePath(node)
	if dsName == "" || dsPath == "" {
		return nil, errors.Forbidden(common.ServiceTree, "Cannot write to root node or to datasource node")
	}

	if ds, ok := s.datasourceByName(dsName); ok {

		node.Path = dsPath
		dsReq := &tree.CreateNodeRequest{
			Node:           node,
			UpdateIfExists: req.UpdateIfExists,
			Silent:         req.Silent,
		}

		response, e := ds.writer.CreateNode(ctx, dsReq)
		if e != nil {
			return nil, e
		}
		s.updateDataSourceNode(response.Node, dsName)
		resp.Node = response.Node

		return resp, nil
	}

	return nil, errors.Forbidden("datasource.not.found", "Unknown datasource %s", dsName)
}

// ReadNode implementation for the TreeServer
func (s *TreeServer) ReadNode(ctx context.Context, req *tree.ReadNodeRequest) (*tree.ReadNodeResponse, error) {

	// Backward compat
	if req.WithExtendedStats {
		req.StatFlags = append(req.StatFlags, tree.StatFlagFolderCounts)
	}
	flags := tree.StatFlags(req.StatFlags)

	node := req.GetNode()
	var metaStreamer meta.Loader
	if flags.Metas() {
		if ms := ctx.Value("MetaStreamer"); ms != nil {
			metaStreamer = ms.(meta.Loader)
		} else {
			metaStreamer = meta.NewStreamLoader(ctx, flags)
			defer metaStreamer.Close()
		}
	}
	resp := &tree.ReadNodeResponse{}
	defer track("ReadNode", ctx, time.Now(), req, resp)

	if node.GetPath() == "" && node.GetUuid() != "" {
		respNode, err := s.lookUpByUuid(ctx, node.GetUuid(), req.StatFlags...)
		if err != nil {
			return nil, err
		}
		resp.Node = respNode
		if flags.Metas() {
			metaStreamer.LoadMetas(ctx, resp.Node)
		}
		return resp, nil
	}

	dsName, dsPath := s.treeNodeToDataSourcePath(node)

	if dsName == "" && dsPath == "" {
		resp.Node = &tree.Node{Uuid: "ROOT", Path: "/"}
		return resp, nil
	}

	if ds, ok := s.datasourceByName(dsName); ok {

		dsReq := &tree.ReadNodeRequest{
			Node:      &tree.Node{Path: dsPath},
			StatFlags: req.StatFlags,
		}

		response, rErr := ds.reader.ReadNode(ctx, dsReq, grpc.WaitForReady(false))
		if rErr != nil {
			e := errors.FromError(rErr)

			return nil, e
		}

		resp.Node = response.Node
		s.updateDataSourceNode(resp.Node, dsName)
		if flags.Metas() {
			metaStreamer.LoadMetas(ctx, resp.Node)
		}

		return resp, nil
	}

	return nil, errors.NotFound(node.GetPath(), "Not found")
}

func (s *TreeServer) ListNodes(req *tree.ListNodesRequest, resp tree.NodeProvider_ListNodesServer) error {

	ctx := resp.Context()
	defer track("ListNodes", ctx, time.Now(), req, resp)

	mainCtx := servicecontext.WithRegistry(ctx, servicecontext.GetRegistry(s.mainCtx))
	var metaStreamer meta.Loader
	var loadMetas bool
	flags := tree.StatFlags(req.StatFlags)
	if flags.Metas() {
		loadMetas = true
		metaStreamer = meta.NewStreamLoader(mainCtx, flags)
		defer metaStreamer.Close()
	}

	// Special case to get ancestors
	if req.Ancestors {

		// FIRST FIND NODE & DS
		var dsName, dsPath string
		sendNode := req.Node

		if req.Node.GetPath() == "" && req.Node.GetUuid() != "" {
			log.Logger(ctx).Debug("First Find node by uuid ", zap.String("uuid", req.Node.GetUuid()))

			var err error
			sendNode, err = s.lookUpByUuid(ctx, req.Node.GetUuid())
			if err != nil {
				return err
			}
			dsName, dsPath = s.treeNodeToDataSourcePath(sendNode)
		} else {
			dsName, dsPath = s.treeNodeToDataSourcePath(req.Node)
		}
		if dsName == "" && dsPath == "" {
			// ROOT NODE
			return errors.BadRequest(common.ServiceTree, "Cannot get ancestors on ROOT node!")

		}

		// Special case for Ancestors: send initial node
		if sendNode.Uuid == "" {
			if dsName != "" && dsPath == "" {
				sendNode.Uuid = "DATASOURCE:" + dsName
			} else {
				// Pass MetaStreamer to avoid spinning a new one
				msCtx := context.WithValue(ctx, "MetaStreamer", metaStreamer)
				if readResp, err := s.ReadNode(msCtx, &tree.ReadNodeRequest{Node: sendNode, StatFlags: req.StatFlags}); err != nil {
					return err
				} else {
					sendNode = readResp.Node
				}
			}
		}

		if loadMetas {
			metaStreamer.LoadMetas(ctx, sendNode)
		}
		resp.Send(&tree.ListNodesResponse{
			Node: sendNode,
		})

		if len(dsPath) > 0 {

			ds, ok := s.datasourceByName(dsName)
			if !ok {
				return errors.BadRequest(common.ServiceTree, "Cannot find datasource client for %s", dsName)
			}
			sendNode.Path = dsPath
			streamer, err := ds.reader.ListNodes(ctx, &tree.ListNodesRequest{
				Node:      sendNode,
				Ancestors: true,
			}, grpc.WaitForReady(false))
			if err != nil {
				return errors.InternalServerError(common.ServiceTree, "Cannot send List request to underlying datasource %s", err.Error())
			}

			defer streamer.CloseSend()
			for {
				listResponse, err := streamer.Recv()
				if listResponse == nil || err != nil {
					break
				}
				respNode := listResponse.Node
				s.updateDataSourceNode(respNode, dsName)
				if loadMetas {
					metaStreamer.LoadMetas(ctx, respNode)
				}
				resp.Send(&tree.ListNodesResponse{
					Node: respNode,
				})
			}

		}

		// NOW SEND ROOT NODE
		resp.Send(&tree.ListNodesResponse{
			Node: &tree.Node{
				Uuid: "ROOT",
				Path: "/",
			},
		})

		return nil

	} else {

		var numberSent, cursorIndex int64
		numberSent, cursorIndex = 0, 0
		return s.ListNodesWithLimit(ctx, metaStreamer, req, resp, &cursorIndex, &numberSent)

	}

}

// ListNodesWithLimit implementation for the TreeServer
func (s *TreeServer) ListNodesWithLimit(ctx context.Context, metaStreamer meta.Loader, req *tree.ListNodesRequest, resp tree.NodeProvider_ListNodesServer, cursorIndex *int64, numberSent *int64) error {

	defer track("ListNodesWithLimit", ctx, time.Now(), req, resp)

	node := req.GetNode()

	dsName, dsPath := s.treeNodeToDataSourcePath(node)
	limit := req.Limit
	offset := req.Offset

	checkLimit := func() bool {
		*numberSent++
		if limit == 0 {
			return false
		}
		if *numberSent >= limit {
			return true
		}
		return false
	}

	if dsName == "" {

		var names []string
		s.sourcesLock.RLock()
		for name := range s.sources {
			names = append(names, name)
		}
		s.sourcesLock.RUnlock()
		log.Logger(ctx).Debug("Should List datasources", zap.Strings("names", names))
		metaFilter := tree.NewMetaFilter(node)
		hasFilter := metaFilter.Parse()
		limitDepth := metaFilter.LimitDepth()

		for _, name := range names {

			if offset > 0 && offset < int64(len(names)) && offset > *cursorIndex {
				*cursorIndex++
				continue
			}
			outputNode := &tree.Node{
				Uuid: "DATASOURCE:" + name,
				Path: name,
			}
			outputNode.MustSetMeta(common.MetaNamespaceNodeName, name)

			ds, _ := s.datasourceByName(name)
			if size, counts, er := s.dsSize(ctx, ds, req.StatFlags); er == nil {
				outputNode.Size = size
				if tree.StatFlags(req.StatFlags).RecursiveCount() {
					outputNode.MustSetMeta(common.MetaFlagRecursiveCount, counts)
				}
			} else {
				log.Logger(ctx).Error("Cannot compute DataSource size, skipping", zap.String("dsName", name), zap.Error(er))
			}
			if req.FilterType == tree.NodeType_UNKNOWN && (!hasFilter || metaFilter.Match(name, outputNode)) {
				if metaStreamer != nil {
					metaStreamer.LoadMetas(ctx, outputNode)
				}
				_ = resp.Send(&tree.ListNodesResponse{
					Node: outputNode,
				})
			}
			*cursorIndex++
			if req.Recursive && limitDepth != 1 {
				subNode := node.Clone()
				subNode.Path = name
				er := s.ListNodesWithLimit(ctx, metaStreamer, &tree.ListNodesRequest{
					Node:         subNode,
					Recursive:    true,
					WithVersions: req.GetWithVersions(),
					StatFlags:    req.GetStatFlags(),
					FilterType:   req.GetFilterType(),
					SortField:    req.GetSortField(),
					SortDirDesc:  req.GetSortDirDesc(),
				}, resp, cursorIndex, numberSent)
				if er != nil {
					return er
				}
			}
			if checkLimit() {
				return nil
			}
		}
		return nil
	}

	if ds, ok := s.datasourceByName(dsName); ok {

		reqNode := node.Clone()
		reqNode.Path = dsPath
		req := &tree.ListNodesRequest{
			Node:      reqNode,
			Recursive: req.Recursive,
			//Limit:      req.Limit,
			StatFlags:   req.GetStatFlags(),
			FilterType:  req.GetFilterType(),
			SortField:   req.GetSortField(),
			SortDirDesc: req.GetSortDirDesc(),
		}

		log.Logger(ctx).Debug("List Nodes With Offset / Limit", zap.Int64("offset", offset), zap.Int64("limit", limit))
		stream, err := ds.reader.ListNodes(ctx, req, grpc.WaitForReady(false))
		if err != nil {
			log.Logger(ctx).Error("ListNodesWithLimit", zap.Error(err))
			return err
		}
		for {
			clientResponse, err := stream.Recv()

			if clientResponse == nil {
				break
			}

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}

			if err != nil {
				return err
			}

			isHidden := strings.HasSuffix(clientResponse.Node.GetPath(), common.PydioSyncHiddenFile)

			if offset > 0 && offset > *cursorIndex {
				if !isHidden {
					*cursorIndex++
				}
				continue
			}

			s.updateDataSourceNode(clientResponse.Node, dsName)
			if metaStreamer != nil {
				metaStreamer.LoadMetas(ctx, clientResponse.Node)
			}
			resp.Send(clientResponse)
			*cursorIndex++

			if !isHidden && checkLimit() {
				return nil
			}
		}

		return nil
	}

	return errors.NotFound(node.GetPath(), "Not found")
}

func (s *TreeServer) dsSize(ctx context.Context, ds DataSource, flags []uint32) (int64, int, error) {
	st, er := ds.reader.ListNodes(ctx, &tree.ListNodesRequest{
		Node:      &tree.Node{Path: ""},
		StatFlags: flags,
	}, grpc.WaitForReady(false))
	if er != nil {
		return 0, 0, er
	}
	var size int64
	var count int
	for {
		if r, e := st.Recv(); e != nil {
			break
		} else {
			size += r.GetNode().GetSize()
			var rc int
			if err := r.GetNode().GetMeta(common.MetaFlagRecursiveCount, &rc); err == nil {
				count += rc
			}
		}
	}
	return size, count, nil
}

// UpdateNode implementation for the TreeServer
func (s *TreeServer) UpdateNode(ctx context.Context, req *tree.UpdateNodeRequest) (*tree.UpdateNodeResponse, error) {

	defer track("UpdateNode", ctx, time.Now(), req, nil)

	from := req.GetFrom()
	to := req.GetTo()

	dsNameFrom, dsPathFrom := s.treeNodeToDataSourcePath(from)
	dsNameTo, dsPathTo := s.treeNodeToDataSourcePath(to)
	if dsNameFrom == "" || dsNameTo == "" || dsPathFrom == "" || dsPathTo == "" {
		return nil, errors.Forbidden(common.ServiceTree, "Cannot write to root node or to datasource node")
	}
	if dsNameFrom != dsNameTo {
		return nil, errors.Forbidden(common.ServiceTree, "Cannot move between two different datasources")
	}

	if ds, ok := s.datasourceByName(dsNameTo); ok {

		from.Path = dsPathFrom
		to.Path = dsPathTo

		req := &tree.UpdateNodeRequest{From: from, To: to}

		if response, er := ds.writer.UpdateNode(ctx, req); er == nil {
			return &tree.UpdateNodeResponse{Success: response.Success, Node: response.Node}, nil
		} else {
			return nil, er
		}

	}

	return nil, errors.Forbidden("datasource.not.found", "Unknown datasource %s", dsNameTo)
}

// DeleteNode implementation for the TreeServer
func (s *TreeServer) DeleteNode(ctx context.Context, req *tree.DeleteNodeRequest) (*tree.DeleteNodeResponse, error) {

	resp := &tree.DeleteNodeResponse{}
	defer track("DeleteNode", ctx, time.Now(), req, resp)

	node := req.GetNode()
	dsName, dsPath := s.treeNodeToDataSourcePath(node)
	if dsName == "" || dsPath == "" {
		return nil, errors.Forbidden(common.ServiceTree, "Cannot delete root node or datasource node")
	}

	if ds, ok := s.datasourceByName(dsName); ok {
		node.Path = dsPath
		if response, e := ds.writer.DeleteNode(ctx, &tree.DeleteNodeRequest{Node: node}); e != nil {
			return nil, e
		} else {
			resp.Success = response.Success
		}
		return resp, nil
	}

	return nil, errors.Forbidden("datasource.not.found", "Unknown datasource %s", dsName)
}

func (s *TreeServer) PublishChange(change *tree.NodeChangeEvent) {
	defer func() {
		if e := recover(); e != nil {
			log.Logger(context.Background()).Error("Panic recovered in PublishChange", zap.Any("error", e))
		}
	}()
	for _, l := range s.listeners {
		if !l.closing {
			l.in <- change
		}
	}
}

func (s *TreeServer) StreamChanges(req *tree.StreamChangesRequest, streamer tree.NodeChangesStreamer_StreamChangesServer) error {

	li := newListener()
	s.listeners = append(s.listeners, li)
	defer func() {
		var cleared []*changesListener
		for _, l := range s.listeners {
			if l == li {
				l.stop()
			} else {
				cleared = append(cleared, l)
			}
		}
		s.listeners = cleared
	}()

	filterPath := strings.Trim(req.RootPath, "/") + "/"

loop:
	for {
		select {
		case <-streamer.Context().Done():
			break loop

		case event := <-li.out:

			if event.Optimistic {
				continue
			}
			newEvent := proto.Clone(event).(*tree.NodeChangeEvent)
			sourceOut := newEvent.Source != nil && !strings.HasPrefix(newEvent.Source.Path, filterPath)
			targetOut := newEvent.Target != nil && !strings.HasPrefix(newEvent.Target.Path, filterPath)
			if (sourceOut && targetOut) || (sourceOut && newEvent.Target == nil) || (targetOut && newEvent.Source == nil) {
				continue
			}
			var scan *tree.Node
			if sourceOut {
				newEvent.Type = tree.NodeChangeEvent_CREATE
				if !event.Target.IsLeaf() {
					scan = event.Target
				}
				newEvent.Source = nil
			} else if targetOut {
				newEvent.Type = tree.NodeChangeEvent_DELETE
				newEvent.Target = nil
			}

			/*
				if newEvent.Target != nil {
					//newEvent.Target.Path = strings.TrimPrefix(newEvent.Target.Path, filterPath)
				}
				if newEvent.Source != nil {
					//newEvent.Source.Path = strings.TrimPrefix(newEvent.Source.Path, filterPath)
				}
			*/
			if newEvent.Metadata != nil {
				// Do not forward this metadata to clients
				delete(newEvent.Metadata, common.XPydioSessionUuid)
				delete(newEvent.Metadata, common.XPydioMoveUuid)
			}
			if e := streamer.Send(newEvent); e != nil {
				return e
			}

			if scan != nil {
				// A folder was move from "outside" to "inside" the filterPath
				// This is a create, and we have to emulate CREATE for all children
				listNodeStreamer := mocks.NewListNodeStreamer()
				wg := sync.WaitGroup{}
				wg.Add(2)
				go func() {
					defer wg.Done()
					s.ListNodes(&tree.ListNodesRequest{Node: scan, Recursive: true}, listNodeStreamer)
				}()
				go func() {
					defer wg.Done()
					defer listNodeStreamer.Close()
					for {
						r, e := listNodeStreamer.Recv()
						if e != nil || r == nil {
							break
						}
						child := r.Node
						//child.Path = strings.TrimPrefix(child.Path, filterPath)
						streamer.Send(&tree.NodeChangeEvent{
							Type:   tree.NodeChangeEvent_CREATE,
							Target: child,
						})
					}
				}()
				wg.Wait()
			}
		}

	}

	return nil
}

func (s *TreeServer) lookUpByUuid(ctx context.Context, uuid string, statFlags ...uint32) (*tree.Node, error) {

	var foundNode *tree.Node

	if strings.HasPrefix(uuid, "DATASOURCE:") {
		dsName := strings.TrimPrefix(uuid, "DATASOURCE:")

		if ds, ok := s.datasourceByName(dsName); ok {
			resp, err := ds.reader.ReadNode(ctx, &tree.ReadNodeRequest{
				Node:      &tree.Node{Uuid: "ROOT"},
				StatFlags: statFlags,
			}, grpc.WaitForReady(false))
			if err == nil && resp.Node != nil {
				s.updateDataSourceNode(resp.Node, dsName)
				log.Logger(ctx).Debug("[Look Up] Found node", zap.String("uuid", resp.Node.Uuid), zap.String("datasource", dsName))
				return resp.Node, nil
			}
		}
		return nil, errors.NotFound(uuid, "Not found")
	}

	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	wg := &sync.WaitGroup{}

	s.sourcesLock.RLock()
	for dsName, ds := range s.sources {
		wg.Add(1)
		reader := ds.reader
		name := dsName
		go func() {
			defer wg.Done()

			resp, err := reader.ReadNode(c, &tree.ReadNodeRequest{
				Node:      &tree.Node{Uuid: uuid},
				StatFlags: statFlags,
			}, grpc.WaitForReady(false))
			if err == nil && resp.Node != nil {
				s.updateDataSourceNode(resp.Node, name)

				log.Logger(ctx).Debug("[Look Up] Found node", zap.String("uuid", resp.Node.Uuid), zap.String("datasource", name))
				foundNode = resp.Node
				cancel()
			}
		}()
	}
	wg.Wait()
	s.sourcesLock.RUnlock()

	if foundNode != nil {
		return foundNode, nil
	} else {
		return nil, errors.NotFound(common.ServiceTree, fmt.Sprintf("Node %s Not found in tree!", uuid))
	}

}

func track(fn string, ctx context.Context, start time.Time, req interface{}, resp interface{}) {
	log.Logger(ctx).Debug(fn, zap.Duration("time", time.Since(start)), zap.Any("req", req), zap.Any("resp", resp))
}
