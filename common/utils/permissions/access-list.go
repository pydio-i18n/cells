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

package permissions

import (
	"context"
	"github.com/pydio/cells/v4/common/utils/std"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/utils/configx"
	json "github.com/pydio/cells/v4/common/utils/jsonx"
)

// PolicyResolver implements the check of an object against a set of ACL policies
type PolicyResolver func(ctx context.Context, request *idm.PolicyEngineRequest, explicitOnly bool) (*idm.PolicyEngineResponse, error)

// VirtualPathResolver must be able to load virtual nodes based on their UUID
type VirtualPathResolver func(context.Context, *tree.Node) (*tree.Node, bool)

// WsLoader is resolver for loading workspaces by their UUIDs
type WsLoader func(ctx context.Context, uuids []string) ([]*idm.Workspace, error)

const (
	FrontWsScopeAll    = "PYDIO_REPO_SCOPE_ALL"
	FrontWsScopeShared = "PYDIO_REPO_SCOPE_SHARED"
)

var (
	AclRead              = &idm.ACLAction{Name: "read", Value: "1"}
	AclWrite             = &idm.ACLAction{Name: "write", Value: "1"}
	AclDeny              = &idm.ACLAction{Name: "deny", Value: "1"}
	AclPolicy            = &idm.ACLAction{Name: "policy"}
	AclQuota             = &idm.ACLAction{Name: "quota"}
	AclLock              = &idm.ACLAction{Name: "lock"}
	AclChildLock         = &idm.ACLAction{Name: "child_lock"}
	AclContentLock       = &idm.ACLAction{Name: "content_lock"}
	AclFrontAction_      = &idm.ACLAction{Name: "action:*"}
	AclFrontParam_       = &idm.ACLAction{Name: "parameter:*"}
	AclWsrootActionName  = "workspace-path"
	AclRecycleRoot       = &idm.ACLAction{Name: "recycle_root", Value: "1"}
	ResolvePolicyRequest PolicyResolver
)

func init() {
	// Use default resolver (loads policies in memory and cache them for 1mn)
	ResolvePolicyRequest = LocalACLPoliciesResolver
}

// AccessList is a merged representation of all ACLs that a user has access to.
// ACLs are merged using a Bitmask form to ease flags detections and comparisons.
type AccessList struct {
	wss           map[string]*idm.Workspace
	wssRootsMasks map[string]map[string]Bitmask
	orderedRoles  []*idm.Role

	wsACLs    []*idm.ACL
	frontACLs []*idm.ACL

	masksByUUIDs  map[string]Bitmask
	masksByPaths  map[string]Bitmask
	maskBULock    *sync.RWMutex
	maskBPLock    *sync.RWMutex
	maskRootsLock *sync.RWMutex
	claimsScopes  map[string]Bitmask

	hasClaimsScopes bool
	cacheKey        string
}

// NewAccessList creates a new AccessList.
func NewAccessList(roles ...*idm.Role) *AccessList {
	acl := &AccessList{
		wss:           make(map[string]*idm.Workspace),
		orderedRoles:  roles,
		maskBPLock:    &sync.RWMutex{},
		maskBULock:    &sync.RWMutex{},
		maskRootsLock: &sync.RWMutex{},
	}
	return acl
}

// AppendACLs appends an additional list of ACLs.
func (a *AccessList) AppendACLs(aa ...*idm.ACL) {
	a.wsACLs = append(a.wsACLs, aa...)
}

// AppendFrontACLs appends an additional list of front-related ACLs.
func (a *AccessList) AppendFrontACLs(aa ...*idm.ACL) {
	a.frontACLs = append(a.frontACLs, aa...)
}

// AppendRoles appends one or more roles. They are kept in order, which is very important.
func (a *AccessList) AppendRoles(rr ...*idm.Role) {
	a.orderedRoles = append(a.orderedRoles, rr...)
}

// GetWorkspaces returns internally stored workspaces
func (a *AccessList) GetWorkspaces() map[string]*idm.Workspace {
	return a.wss
}

// LoadWorkspaces loads actual idm.Workspace objects using a WsLoader
func (a *AccessList) LoadWorkspaces(ctx context.Context, loader WsLoader) error {
	a.maskRootsLock.RLock()
	defer a.maskRootsLock.RUnlock()
	if len(a.wssRootsMasks) == 0 {
		// nothing to do
		return nil
	}
	var uuids []string
	for wsId := range a.wssRootsMasks {
		uuids = append(uuids, wsId)
	}
	ww, er := loader(ctx, uuids)
	if er != nil {
		return er
	}
	for _, w := range ww {
		// Set rootId on workspace
		for rootId := range a.wssRootsMasks[w.UUID] {
			w.RootUUIDs = append(w.RootUUIDs, rootId)
		}
		a.wss[w.UUID] = w
	}
	return nil
}

// GetRoles returns ordered list of roles
func (a *AccessList) GetRoles() []*idm.Role {
	return a.orderedRoles
}

// AppendClaimsScopes appends some specific permissions passed through claims.
// Currently only strings like "node:uuid:perm" are supported
func (a *AccessList) AppendClaimsScopes(ss []string) {
	a.hasClaimsScopes = true
	a.parseClaimScopes(ss)
}

// HasPolicyBasedAcls checks if there are policy based acls.
func (a *AccessList) HasPolicyBasedAcls() bool {
	for _, acl := range a.wsACLs {
		if acl.Action.Name == AclPolicy.Name {
			return true
		}
	}
	return false
}

// Flatten performs actual flatten.
func (a *AccessList) Flatten(ctx context.Context) {
	nodes, workspaces := a.flattenNodes(ctx, a.wsACLs)
	a.maskBULock.Lock()
	a.masksByUUIDs = nodes
	a.maskBULock.Unlock()
	a.wssRootsMasks = workspaces
}

// GetWorkspacesRoots gets detected workspace root nodes that are then
// used to populate the Workspace keys.
func (a *AccessList) GetWorkspacesRoots() map[string]map[string]Bitmask {
	a.maskRootsLock.RLock()
	cl := std.CloneMap(a.wssRootsMasks)
	a.maskRootsLock.RUnlock()
	return cl
}

// GetNodesBitmasks returns internal bitmask
func (a *AccessList) GetNodesBitmasks() map[string]Bitmask {
	a.maskBULock.RLock()
	defer a.maskBULock.RUnlock()
	cp := make(map[string]Bitmask, len(a.masksByUUIDs))
	for k, v := range a.masksByUUIDs {
		cp[k] = v
	}
	return cp
}

// AddNodeBitmask appends a node bitmask to the internal list
func (a *AccessList) AddNodeBitmask(id string, b Bitmask) {
	a.maskBULock.Lock()
	a.masksByUUIDs[id] = b
	a.maskBULock.Unlock()
}

// ReplicateBitmask copies a bitmask value from one position to another
func (a *AccessList) ReplicateBitmask(fromUuid, toUuid string, replaceInRoots ...bool) bool {
	replace := len(replaceInRoots) > 0 && replaceInRoots[0]
	// Protect this method from concurrency
	a.maskBULock.Lock()
	defer func() {
		a.maskBULock.Unlock()
		if replace && a.cacheKey != "" {
			//fmt.Println("Updating acl in cache with key", a.cacheKey)
			_ = a.cache(a.cacheKey)
		}
	}()
	if b, o := a.masksByUUIDs[fromUuid]; o {
		a.masksByUUIDs[toUuid] = b
		if replace {
			a.maskRootsLock.Lock()
			// Replace in wssRootsMasks
			for maskId, roots := range a.wssRootsMasks {
				nr := make(map[string]Bitmask)
				modified := false
				for rootId, current := range roots {
					if rootId == fromUuid {
						nr[toUuid] = b
						modified = true
					} else {
						nr[rootId] = current
					}
				}
				if modified {
					a.wssRootsMasks[maskId] = nr
				}
			}
			a.maskRootsLock.Unlock()
		}
		return true
	}
	return false
}

// DetectedWsRights retrieves a map of accessible workspaces.
func (a *AccessList) DetectedWsRights(ctx context.Context) map[string]SimpleRight {
	a.maskRootsLock.RLock()
	defer a.maskRootsLock.RUnlock()

	results := make(map[string]SimpleRight)
	for wsId, wsNodes := range a.wssRootsMasks {
		rights := SimpleRight{}
		for nodeId := range wsNodes {
			if a.CanRead(ctx, &tree.Node{Uuid: nodeId}) {
				rights.Read = true
			}
			if a.CanWrite(ctx, &tree.Node{Uuid: nodeId}) {
				rights.Write = true
			}
		}
		if rights.IsAccessible() {
			results[wsId] = rights
		}
	}
	return results
}

// CanRead checks if a node has READ access.
func (a *AccessList) CanRead(ctx context.Context, nodes ...*tree.Node) bool {
	if len(nodes) == 0 {
		return false
	}
	if a.claimsScopesDeny(ctx, nodes[0], FlagRead) {
		return false
	}
	deny, mask := a.parentMaskOrDeny(ctx, false, nodes...)
	return !deny && mask.HasFlag(ctx, FlagRead, nodes...)
}

// CanWrite checks if a node has WRITE access.
func (a *AccessList) CanWrite(ctx context.Context, nodes ...*tree.Node) bool {
	if len(nodes) == 0 {
		return false
	}
	if a.claimsScopesDeny(ctx, nodes[0], FlagWrite) {
		return false
	}
	deny, mask := a.parentMaskOrDeny(ctx, false, nodes...)
	return !deny && mask.HasFlag(ctx, FlagWrite, nodes...)
}

func (a *AccessList) HasExplicitDeny(ctx context.Context, flag BitmaskFlag, nodes ...*tree.Node) bool {
	_, mask := a.parentMaskOrDeny(ctx, false, nodes...)
	// Only test first node - do not test parents
	if len(nodes) > 1 {
		nodes = nodes[:1]
	}
	return mask.HasPolicyExplicitDeny(ctx, flag, nodes...)
}

// CanReadWithResolver checks if a node has READ access, using VirtualPathResolver if necessary
func (a *AccessList) CanReadWithResolver(ctx context.Context, resolver VirtualPathResolver, nodes ...*tree.Node) bool {
	if len(nodes) == 0 {
		return false
	}
	a.replicateMasksResolved(ctx, resolver)
	if a.claimsScopesDeny(ctx, nodes[0], FlagRead) {
		return false
	}
	deny, mask := a.parentMaskOrDeny(ctx, false, nodes...)
	return !deny && mask.HasFlag(ctx, FlagRead, nodes...)
}

// CanWriteWithResolver checks if a node has WRITE access, using VirtualPathResolver if necessary.
func (a *AccessList) CanWriteWithResolver(ctx context.Context, resolver VirtualPathResolver, nodes ...*tree.Node) bool {
	if len(nodes) == 0 {
		return false
	}
	a.replicateMasksResolved(ctx, resolver)
	if a.claimsScopesDeny(ctx, nodes[0], FlagWrite) {
		return false
	}
	deny, mask := a.parentMaskOrDeny(ctx, false, nodes...)
	return !deny && mask.HasFlag(ctx, FlagWrite, nodes...)
}

// CanReadPath checks if a node has READ access based on its Path
func (a *AccessList) CanReadPath(ctx context.Context, resolver VirtualPathResolver, nodes ...*tree.Node) bool {
	if a.masksByPaths == nil {
		if e := a.loadNodePathAcls(ctx, resolver); e != nil {
			log.Logger(ctx).Error("Could not load NodePathsAcls", zap.Error(e))
			return false
		}
	}
	deny, mask := a.parentMaskOrDeny(ctx, true, nodes...)
	return !deny && mask.HasFlag(ctx, FlagRead, nodes...)
}

// CanWritePath checks if a node has WRITE access based on its path.
func (a *AccessList) CanWritePath(ctx context.Context, resolver VirtualPathResolver, nodes ...*tree.Node) bool {
	if a.masksByPaths == nil {
		if e := a.loadNodePathAcls(ctx, resolver); e != nil {
			log.Logger(ctx).Error("Could not load NodePathsAcls", zap.Error(e))
			return false
		}
	}
	deny, mask := a.parentMaskOrDeny(ctx, true, nodes...)
	return !deny && mask.HasFlag(ctx, FlagWrite, nodes...)
}

// IsLocked checks if a node bitmask has a FlagLock value.
func (a *AccessList) IsLocked(ctx context.Context, nodes ...*tree.Node) bool {
	// First we check for parents
	if len(nodes) == 0 {
		return false
	}
	mask, _ := a.firstMaskForParents(ctx, nodes...)
	if mask.HasFlag(ctx, FlagLock, nodes[0]) {
		return true
	}

	if mask := a.firstMaskForChildren(ctx, nodes[0]); mask.HasFlag(ctx, FlagLock, nodes[0]) {
		return true
	}

	return false
}

// BelongsToWorkspaces finds corresponding workspace parents for this node.
func (a *AccessList) BelongsToWorkspaces(ctx context.Context, nodes ...*tree.Node) (workspaces []*idm.Workspace, workspacesRoots map[string]string) {
	a.maskRootsLock.RLock()
	defer a.maskRootsLock.RUnlock()

	foundWorkspaces := make(map[string]bool)
	workspacesRoots = make(map[string]string)
	for _, node := range nodes {
		uuid := node.Uuid
		for wsId, wsRoots := range a.wssRootsMasks {
			if _, has := a.wss[wsId]; !has {
				continue
			}
			for rootId := range wsRoots {
				if rootId == uuid {
					foundWorkspaces[wsId] = true
					workspacesRoots[wsId] = rootId
				}
			}
		}
	}
	for workspaceId := range foundWorkspaces {
		workspaces = append(workspaces, a.wss[workspaceId])
	}
	return workspaces, workspacesRoots

}

// FlattenedFrontValues generates a configx.Values with frontend actions/parameters configs
func (a *AccessList) FlattenedFrontValues() configx.Values {
	output := configx.New()
	for _, role := range a.orderedRoles {
		for _, acl := range a.frontACLs {
			if acl.RoleID != role.Uuid {
				continue
			}
			name := acl.Action.Name
			value := acl.Action.Value
			scope := acl.WorkspaceID
			var iVal interface{}
			if e := json.Unmarshal([]byte(value), &iVal); e != nil {
				// May not be marshalled, use original string instead
				iVal = value
			}
			parts := strings.Split(name, ":")
			t := parts[0]
			p := parts[1]
			n := parts[2]

			if t == "action" {
				output.Val("actions", p, n, scope).Set(iVal)
			} else {
				output.Val("parameters", p, n, scope).Set(iVal)
			}
		}
	}

	return output
}

// Zap simply returns a zapcore.Field object populated with this aggregated AccessList under a standard key
func (a *AccessList) Zap() zapcore.Field {
	return zap.Any(common.KeyAccessList, a)
}

/***************
PRIVATE METHODS
****************/

// loadNodePathAcls retrieve each node by UUID, to which an ACL is attached
func (a *AccessList) loadNodePathAcls(ctx context.Context, resolver VirtualPathResolver) error {
	a.maskBULock.RLock()
	defer a.maskBULock.RUnlock()
	a.maskBPLock.Lock()
	defer a.maskBPLock.Unlock()

	a.masksByPaths = make(map[string]Bitmask, len(a.masksByUUIDs))
	if len(a.masksByUUIDs) == 0 {
		// Do not open an unnecessary stream...
		return nil
	}
	cli := tree.NewNodeProviderStreamerClient(grpc.GetClientConnFromCtx(ctx, common.ServiceTree))
	ct, ca := context.WithCancel(ctx)
	defer ca()
	st, e := cli.ReadNodeStream(ct)
	if e != nil {
		return e
	}
	// Retrieving path foreach ids
	for nodeID, b := range a.masksByUUIDs {
		if n, ok := resolver(ctx, &tree.Node{Uuid: nodeID}); ok {
			log.Logger(ctx).Debug("Acl.loadNodePathAcls : Loading resolved node", n.Zap())
			a.masksByPaths[strings.TrimSuffix(n.Path, "/")] = b
			continue
		}
		err := st.Send(&tree.ReadNodeRequest{Node: &tree.Node{Uuid: nodeID}})
		if err != nil {
			return err
		}
		resp, err := st.Recv()
		if err != nil || resp.Node == nil {
			continue
		}
		a.masksByPaths[strings.TrimSuffix(resp.Node.Path, "/")] = b
	}
	return nil
}

// Flatten Permissions based on all the lists received :
// First go through each node and create Bitmask for each one, organized by roles
// Then flatten roles by keeping only last Bitmask found if node appears many times in many roles.
//
// At the same time, collect nodes that are flagged with a "WorkspaceID" in ACL to compute the list of
// accessible roots. Other permissions are used a simple folder permissions, they do not trigger a new workspace
// in the list.
func (a *AccessList) flattenNodes(ctx context.Context, aclList []*idm.ACL) (map[string]Bitmask, map[string]map[string]Bitmask) {

	flattenedNodes := make(map[string]Bitmask)
	flattenedWorkspaces := make(map[string]map[string]Bitmask)

	detectedWorkspaces := make(map[string]map[string]string)
	roles := make(map[string]map[string]Bitmask)

	// Create Bitmasks for each node, for each role
	for _, acl := range aclList {
		if acl.NodeID == "" {
			continue
		}
		var roleNodes map[string]Bitmask
		if test, ok := roles[acl.RoleID]; ok {
			roleNodes = test
		} else {
			roleNodes = make(map[string]Bitmask)
		}

		var nodeAcls Bitmask
		if test, ok := roleNodes[acl.NodeID]; ok {
			nodeAcls = test
		} else {
			nodeAcls = Bitmask{}
		}
		if flag, ok := NamesToFlags[acl.Action.Name]; ok {
			if flag == FlagPolicy {
				nodeAcls.AddPolicyFlag(acl.Action.Value)
			} else if flag == FlagQuota {
				nodeAcls.AddValueFlag(flag, acl.Action.Value)
			} else {
				nodeAcls.AddFlag(flag)
			}
		}

		roleNodes[acl.NodeID] = nodeAcls
		roles[acl.RoleID] = roleNodes

		if acl.WorkspaceID != "" {
			if _, ok := detectedWorkspaces[acl.WorkspaceID]; !ok {
				detectedWorkspaces[acl.WorkspaceID] = make(map[string]string)
			}
			detectedWorkspaces[acl.WorkspaceID][acl.NodeID] = acl.NodeID
		}
	}

	// Now flatten on roles : last role wins on each node
	for _, role := range a.orderedRoles {
		if roleNodes, ok := roles[role.Uuid]; ok {
			for nodeId, bitmask := range roleNodes {
				flattenedNodes[nodeId] = bitmask
			}
		}
	}
	for workspaceId, workspaceRootNodes := range detectedWorkspaces {
		wsRoots := make(map[string]Bitmask)
		for nodeId := range workspaceRootNodes {
			mask := flattenedNodes[nodeId]
			if !mask.HasFlag(ctx, FlagDeny) {
				wsRoots[nodeId] = mask
			}
		}
		if len(wsRoots) > 0 {
			flattenedWorkspaces[workspaceId] = wsRoots
		}
	}

	return flattenedNodes, flattenedWorkspaces
}

// replicateMasksResolved creates resolves internal masksByUUIDs by UUID using passed resolver
func (a *AccessList) replicateMasksResolved(ctx context.Context, resolver VirtualPathResolver) {
	// ReplicateBitmask takes the lock - do not use it with defer here. Just copy ids to a slice.
	var ids []string
	a.maskBULock.RLock()
	for id := range a.masksByUUIDs {
		ids = append(ids, id)
	}
	a.maskBULock.RUnlock()
	for _, id := range ids {
		if res, o := resolver(ctx, &tree.Node{Uuid: id}); o {
			a.ReplicateBitmask(id, res.Uuid)
		}
	}
}

// parentMaskOrDeny browses access list from current node to ROOT, going through each parent.
// If there is a DENY anywhere up the path, it returns that deny,
// otherwise it sends the first Bitmask found (the closest parent having a Bitmask set).
func (a *AccessList) parentMaskOrDeny(ctx context.Context, byPath bool, nodes ...*tree.Node) (bool, Bitmask) {
	var firstParent Bitmask
	var hasParentDeny bool
	var checkOn map[string]Bitmask
	var locker *sync.RWMutex
	if byPath {
		checkOn = a.masksByPaths
		locker = a.maskBPLock
	} else {
		checkOn = a.masksByUUIDs
		locker = a.maskBULock
	}
	locker.RLock()
	defer locker.RUnlock()

	for _, node := range nodes {
		var checkKey string
		if byPath {
			checkKey = strings.Trim(node.Path, "/")
		} else {
			checkKey = node.Uuid
		}
		if bitmask, ok := checkOn[checkKey]; ok {
			if firstParent.BitmaskFlag == BitmaskFlag(0) {
				firstParent = bitmask
			}
			if bitmask.HasFlag(ctx, FlagDeny, node) {
				return true, Bitmask{BitmaskFlag: FlagDeny}
			}
		}
	}
	return hasParentDeny, firstParent
}

// firstMaskForParents just climbs up the tree and gets the first non-empty mask found.
func (a *AccessList) firstMaskForParents(ctx context.Context, nodes ...*tree.Node) (Bitmask, *tree.Node) {
	a.maskBULock.RLock()
	defer a.maskBULock.RUnlock()

	for _, node := range nodes {
		if bitmask, ok := a.masksByUUIDs[node.Uuid]; ok {
			return bitmask, node
		}
	}
	return Bitmask{}, nil
}

// firstMaskForChildren look through all the access list pathes to get the first mask available for the node given in argument
func (a *AccessList) firstMaskForChildren(ctx context.Context, node *tree.Node) Bitmask {
	a.maskBPLock.RLock()
	defer a.maskBPLock.RUnlock()

	keys := make([]string, 0, len(a.masksByPaths))
	for k := range a.masksByPaths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, p := range keys {
		if strings.HasPrefix(p, strings.TrimRight(node.Path, "/")+"/") {
			return a.masksByPaths[p]
		}
	}
	return Bitmask{}
}

// parseClaimScopes parse scopes and store them internally
func (a *AccessList) parseClaimScopes(ss []string) {
	if a.claimsScopes == nil {
		a.claimsScopes = make(map[string]Bitmask)
	}
	for _, s := range ss {
		// Look for scopes like "node:uuid:perm"
		parts := strings.Split(s, ":")
		if len(parts) != 3 || parts[0] != "node" {
			continue
		}
		uuid := parts[1]
		flag := Bitmask{}
		if strings.Contains(parts[2], "r") {
			flag.AddFlag(FlagRead)
		}
		if strings.Contains(parts[2], "w") {
			flag.AddFlag(FlagWrite)
		}
		a.claimsScopes[uuid] = flag
	}
}

// claimsScopesDeny checks if claimsScopes are set and verify node UUID against them
func (a *AccessList) claimsScopesDeny(ctx context.Context, node *tree.Node, perm BitmaskFlag) bool {
	if a.hasClaimsScopes {
		if flag, o := a.claimsScopes[node.Uuid]; !o || !flag.HasFlag(ctx, perm) {
			return true
		}
	}
	return false
}

// SimpleRight is a tool struct to compute SimpleRight strings
type SimpleRight struct {
	Read  bool
	Write bool
}

func (r *SimpleRight) IsAccessible() bool {
	return r.Read || r.Write
}

func (r *SimpleRight) String() string {
	var s []string
	if r.Read {
		s = append(s, "read")
	}
	if r.Write {
		s = append(s, "write")
	}
	return strings.Join(s, ",")
}

func (r *SimpleRight) UserStateString() string {
	var s []string
	if r.Read {
		s = append(s, "r")
	}
	if r.Write {
		s = append(s, "w")
	}
	return strings.Join(s, "")
}
