/*
 * Copyright (c) 2018-2023. Abstrium SAS <team (at) pydio.com>
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

package share

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/auth"
	"github.com/pydio/cells/v4/common/auth/claim"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/rest"
	service2 "github.com/pydio/cells/v4/common/proto/service"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/service/resources"
	"github.com/pydio/cells/v4/common/utils/permissions"
)

type SharedResource struct {
	Node       *tree.Node
	Workspaces []*idm.Workspace
}

// ListSharedResources lists all links and cells Owned by a given user
func (sc *Client) ListSharedResources(ctx context.Context, subject string, scope idm.WorkspaceScope, ownedBy bool, p resources.ResourceProviderHandler) ([]*SharedResource, error) {
	var out []*SharedResource

	var subjects []string
	var userId string
	var admin bool
	if claims, ok := ctx.Value(claim.ContextKey).(claim.Claims); ok {
		admin = claims.Profile == common.PydioProfileAdmin
		userId = claims.Subject
	}
	if subject != "" {
		if !admin {
			return nil, fmt.Errorf("only admins can specify a subject")
		}
		subjects = append(subjects, subject)
	} else {
		var e error
		if subjects, e = auth.SubjectsForResourcePolicyQuery(ctx, &rest.ResourcePolicyQuery{Type: rest.ResourcePolicyQuery_CONTEXT}); e != nil {
			return nil, e
		}
	}

	var qs []*anypb.Any
	if scope == idm.WorkspaceScope_ROOM || scope == idm.WorkspaceScope_ANY {
		q, _ := anypb.New(&idm.WorkspaceSingleQuery{Scope: idm.WorkspaceScope_ROOM})
		qs = append(qs, q)
	}
	if scope == idm.WorkspaceScope_LINK || scope == idm.WorkspaceScope_ANY {
		q, _ := anypb.New(&idm.WorkspaceSingleQuery{Scope: idm.WorkspaceScope_LINK})
		qs = append(qs, q)
	}

	cl := idm.NewWorkspaceServiceClient(grpc.GetClientConnFromCtx(sc.RuntimeContext, common.ServiceWorkspace))
	streamer, err := cl.SearchWorkspace(ctx, &idm.SearchWorkspaceRequest{
		Query: &service2.Query{
			SubQueries: qs,
			Operation:  service2.OperationType_OR,
			ResourcePolicyQuery: &service2.ResourcePolicyQuery{
				Subjects: subjects,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	workspaces := map[string]*idm.Workspace{}
	var workspaceIds []string
	for {
		resp, e := streamer.Recv()
		if e != nil {
			break
		}
		if ownedBy && !p.MatchPolicies(ctx, resp.Workspace.UUID, resp.Workspace.Policies, service2.ResourcePolicyAction_OWNER, userId) {
			continue
		}
		workspaces[resp.Workspace.UUID] = resp.Workspace
		workspaceIds = append(workspaceIds, resp.Workspace.UUID)
	}

	if len(workspaces) == 0 {
		return out, nil
	}

	acls, e := permissions.GetACLsForWorkspace(ctx, workspaceIds, permissions.AclRead, permissions.AclWrite, permissions.AclPolicy)
	if e != nil {
		return nil, e
	}

	// Map roots to objects
	roots := make(map[string]map[string]*idm.Workspace)
	var detectedRoots []string
	for _, acl := range acls {
		if acl.NodeID == "" {
			continue
		}
		if _, has := roots[acl.NodeID]; !has {
			roots[acl.NodeID] = make(map[string]*idm.Workspace)
			detectedRoots = append(detectedRoots, acl.NodeID)
		}
		if ws, ok := workspaces[acl.WorkspaceID]; ok {
			roots[acl.NodeID][acl.WorkspaceID] = ws
		}
	}
	var rootNodes map[string]*tree.Node
	if subject != "" {
		rootNodes = sc.LoadAdminRootNodes(ctx, detectedRoots)
	} else {
		accessList, _ := permissions.AccessListFromContextClaims(ctx)
		rootNodes = sc.LoadDetectedRootNodes(ctx, detectedRoots, accessList)
	}

	// Build resources
	for nodeId, node := range rootNodes {

		resource := &SharedResource{
			Node: node,
		}
		for _, ws := range roots[nodeId] {
			resource.Workspaces = append(resource.Workspaces, ws)
		}
		out = append(out, resource)
	}

	return out, nil

}
