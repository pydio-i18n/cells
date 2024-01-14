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

// Package rest provides a REST service for querying the search engine
package rest

import (
	"context"
	"github.com/pydio/cells/v4/common/service/resources"
	"github.com/pydio/cells/v4/idm/share"
	"regexp"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/pydio/cells/v4/common/client/grpc"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/acl"
	"github.com/pydio/cells/v4/common/nodes/compose"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/rest"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/service"
	"github.com/pydio/cells/v4/common/service/context/metadata"
)

type Handler struct {
	runtimeCtx context.Context
	router     nodes.Client
	client     tree.SearcherClient
}

// SwaggerTags list the names of the service tags declared in the swagger json implemented by this service
func (s *Handler) SwaggerTags() []string {
	return []string{"SearchService"}
}

// Filter returns a function to filter the swagger path
func (s *Handler) Filter() func(string) string {
	return nil
}

func (s *Handler) getRouter() nodes.Client {
	if s.router == nil {
		s.router = compose.PathClient(s.runtimeCtx)
	}
	return s.router
}

func (s *Handler) getClient() tree.SearcherClient {
	if s.client == nil {
		s.client = tree.NewSearcherClient(grpc.GetClientConnFromCtx(s.runtimeCtx, common.ServiceSearch))
	}
	return s.client
}

func (s *Handler) sharedResourcesAsNodes(ctx context.Context, query *tree.Query) ([]*tree.Node, bool, error) {
	scope, freeString, active := s.extractSharedMeta(query.FreeString)
	if !active {
		return nil, false, nil
	}
	// Replace FS
	query.FreeString = freeString

	sc := share.NewClient(s.runtimeCtx, nil)
	rr, e := sc.ListSharedResources(ctx, "", scope, true, resources.ResourceProviderHandler{})
	if e != nil {
		return nil, false, e
	}
	var out []*tree.Node
	for _, r := range rr {
		out = append(out, r.Node)
	}
	return out, active, nil
}

func (s *Handler) extractSharedMeta(freeString string) (scope idm.WorkspaceScope, newString string, has bool) {
	rx, _ := regexp.Compile("((\\+)?Meta\\.shared_resource_type:(any|cell|link))")
	matches := rx.FindAllStringSubmatch(freeString, -1)
	if len(matches) == 1 && len(matches[0]) == 4 {
		has = true
		switch matches[0][3] {
		case "any":
			scope = idm.WorkspaceScope_ANY
		case "cell":
			scope = idm.WorkspaceScope_ROOM
		case "link":
			scope = idm.WorkspaceScope_LINK
		}
		newString = rx.ReplaceAllString(freeString, "")
		newString = strings.TrimSpace(newString)
	}
	return
}

func (s *Handler) Nodes(req *restful.Request, rsp *restful.Response) {

	ctx := req.Request.Context()
	var searchRequest tree.SearchRequest
	if err := req.ReadEntity(&searchRequest); err != nil {
		service.RestError500(req, rsp, err)
		return
	}

	query := searchRequest.Query
	if query == nil {
		rsp.WriteEntity(&rest.SearchResults{Total: 0})
		return
	}

	router := s.getRouter()

	var nn []*tree.Node
	var facets []*tree.SearchFacet
	prefixes := []string{}
	nodesPrefixes := map[string]string{}
	var passedPrefix string
	var passedWorkspaceSlug string
	if len(query.PathPrefix) > 0 {
		passedPrefix = strings.Trim(query.PathPrefix[0], "/")
		if len(strings.Split(passedPrefix, "/")) == 1 {
			passedWorkspaceSlug = passedPrefix
			passedPrefix = ""
		}
	}

	cl := tree.NewNodeProviderStreamerClient(grpc.GetClientConnFromCtx(ctx, common.ServiceTree))
	readCtx := metadata.WithAdditionalMetadata(ctx, tree.StatFlags(searchRequest.StatFlags).AsMeta())
	nodeStreamer, e := cl.ReadNodeStream(readCtx)
	if e == nil {
		defer nodeStreamer.CloseSend()
	}

	// TMP Load shared
	sharedNodes, shared, er := s.sharedResourcesAsNodes(ctx, query)
	if er != nil {
		log.Logger(ctx).Error("cannot load shared resources")
		service.RestErrorDetect(req, rsp, e)
		return
	}

	err := router.WrapCallback(func(inputFilter nodes.FilterFunc, outputFilter nodes.FilterFunc) error {

		var userWorkspaces map[string]*idm.Workspace
		// Fill a context with current user info
		// (Let inputFilter apply the various necessary middlewares).
		loaderCtx, _, _ := inputFilter(ctx, &tree.Node{Path: ""}, "tmp")
		if accessList, ok := acl.FromContext(loaderCtx); ok {
			userWorkspaces = accessList.GetWorkspaces()
		}

		if shared {
			if len(sharedNodes) == 0 {
				// Break now, return an empty result
				return nil
			}
			for _, n := range sharedNodes {
				p := n.GetPath()
				ctx, n, e = inputFilter(ctx, n, "search-"+p)
				if e != nil {
					continue
				}
				log.Logger(ctx).Debug("Filtered Node & Context", zap.String("path", n.Path))
				query.Paths = append(query.Paths, n.Path)
			}
		}

		if len(passedPrefix) > 0 {
			// Passed prefix
			prefixes = append(prefixes, passedPrefix)

		} else {
			for _, w := range userWorkspaces {
				if len(passedWorkspaceSlug) > 0 && w.Slug != passedWorkspaceSlug {
					continue
				}
				if len(w.RootUUIDs) > 1 {
					for _, root := range w.RootUUIDs {
						prefixes = append(prefixes, w.Slug+"/"+root)
					}
				} else {
					prefixes = append(prefixes, w.Slug)
				}
			}
		}
		query.PathPrefix = []string{}

		var e error
		ctx = acl.WithPresetACL(loaderCtx, nil) // Just set the key, acl is already set
		for _, p := range prefixes {
			rootNode := &tree.Node{Path: p}
			ctx, rootNode, e = inputFilter(ctx, rootNode, "search-"+p)
			if e != nil {
				continue
			}
			log.Logger(ctx).Debug("Filtered Node & Context", zap.String("path", rootNode.Path))
			nodesPrefixes[rootNode.Path] = p
			query.PathPrefix = append(query.PathPrefix, rootNode.Path)
		}

		sClient, err := s.getClient().Search(ctx, &searchRequest)
		if err != nil {
			return err
		}

		defer sClient.CloseSend()

		for {
			resp, rErr := sClient.Recv()
			if resp == nil {
				break
			} else if rErr != nil {
				return err
			}
			if resp.Facet != nil {
				facets = append(facets, resp.Facet)
				continue
			}
			respNode := resp.Node
			wrapperCtx, wrapperN, _ := inputFilter(ctx, respNode, "in-"+respNode.Uuid)
			if err := router.WrappedCanApply(wrapperCtx, wrapperCtx, &tree.NodeChangeEvent{Type: tree.NodeChangeEvent_READ, Source: wrapperN}); err != nil {
				log.Logger(ctx).Debug("Skipping node in search results", respNode.ZapPath(), zap.Error(err))
				continue
			}
			for r, p := range nodesPrefixes {
				if strings.HasPrefix(respNode.Path, r+"/") {
					log.Logger(ctx).Debug("Response", zap.String("node", respNode.Path))
					if nodeStreamer != nil {
						nodeStreamer.Send(&tree.ReadNodeRequest{Node: respNode.Clone()})
						if nsR, e := nodeStreamer.Recv(); e == nil {
							respNode = nsR.GetNode()
						}
					}
					_, filtered, err := outputFilter(ctx, respNode, "search-"+p)
					if err != nil {
						return err
					}
					if userWorkspaces != nil {
						for _, w := range userWorkspaces {
							if strings.HasPrefix(filtered.Path, w.Slug+"/") {
								filtered.MustSetMeta(common.MetaFlagWorkspaceRepoId, w.UUID)
								filtered.MustSetMeta(common.MetaFlagWorkspaceRepoDisplay, w.Label)
							}
						}
					}
					nn = append(nn, filtered.WithoutReservedMetas())
				}
			}
		}
		return nil

	})

	if err != nil {
		log.Logger(ctx).Error("Query", zap.Error(err))
		service.RestError500(req, rsp, err)
		return
	}

	result := &rest.SearchResults{
		Results: nn,
		Facets:  facets,
		Total:   int32(len(nn)),
	}
	rsp.WriteEntity(result)

}
