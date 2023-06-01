package grpc

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	service "github.com/pydio/cells/v4/common/proto/service"
	"github.com/pydio/cells/v4/common/proto/tree"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/utils/permissions"
	"github.com/pydio/cells/v4/common/utils/std"
	"github.com/pydio/cells/v4/idm/acl"
)

// UpgradeTo120 looks for workspace roots and CellNode roots and set a "recycle_root" flag on them.
func UpgradeTo120(ctx context.Context) error {

	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	// REMOVE pydiogateway ACLs
	log.Logger(ctx).Info("ACLS: remove pydiogateway ACLs")
	q1, _ := anypb.New(&idm.ACLSingleQuery{
		WorkspaceIDs: []string{"pydiogateway"},
	})
	if num, e := dao.Del(&service.Query{SubQueries: []*anypb.Any{q1}}, nil); e != nil {
		log.Logger(ctx).Error("Could not delete pydiogateway acls, please manually remove them from ACLs!", zap.Error(e))
	} else {
		log.Logger(ctx).Info("Removed pydiogateway acls", zap.Int64("numRows", num))
	}

	// ADD recycle_root on workspaces
	log.Logger(ctx).Info("Upgrading ACLs for recycle_root flags")
	metaClient := tree.NewNodeProviderClient(grpc.GetClientConnFromCtx(ctx, common.ServiceMeta))
	q, _ := anypb.New(&idm.ACLSingleQuery{
		Actions: []*idm.ACLAction{
			{Name: permissions.AclWsrootActionName},
		},
	})
	acls := new([]interface{})
	dao.Search(&service.Query{
		SubQueries: []*anypb.Any{q},
	}, acls, nil)
	for _, in := range *acls {
		val, ok := in.(*idm.ACL)
		if !ok {
			continue
		}
		var addRecycleAcl bool
		wsPath := val.Action.Value
		if strings.HasPrefix(wsPath, "uuid:") {
			// Load meta from node Uuid and check if it has a CellNode flag
			nodeUuid := strings.TrimPrefix(wsPath, "uuid:")
			std.Retry(ctx, func() error {
				log.Logger(ctx).Info("Loading metadata for node to check if it's a CellNode root")
				if r, e := metaClient.ReadNode(ctx, &tree.ReadNodeRequest{Node: &tree.Node{Uuid: nodeUuid}}); e == nil && r.Node.GetMetaBool(common.MetaFlagCellNode) {
					addRecycleAcl = true
				}
				return nil
			}, 4*time.Second)
		} else {
			addRecycleAcl = true
		}
		if addRecycleAcl {
			newAcl := &idm.ACL{
				WorkspaceID: val.WorkspaceID,
				NodeID:      val.NodeID,
				RoleID:      val.RoleID,
				Action:      permissions.AclRecycleRoot,
			}
			log.Logger(ctx).Info("Inserting new ACL")
			if e := dao.Add(newAcl); e != nil {
				log.Logger(ctx).Error("-- Could not create recycle_root ACL", zap.Error(e))
			}
		}
	}

	treeClient := tree.NewNodeProviderClient(grpc.GetClientConnFromCtx(ctx, common.ServiceTree))
	// Special case for personal files: browse existing folders, assume they are users personal workspaces and add recycle root
	std.Retry(ctx, func() error {
		stream, e := treeClient.ListNodes(ctx, &tree.ListNodesRequest{Node: &tree.Node{Path: "personal"}})
		if e != nil {
			return e
		}
		defer stream.CloseSend()
		for {
			resp, er := stream.Recv()
			if er != nil {
				break
			}
			if resp == nil || resp.Node == nil || resp.Node.IsLeaf() {
				continue
			}
			newAcl := &idm.ACL{
				NodeID: resp.Node.Uuid,
				Action: permissions.AclRecycleRoot,
			}
			log.Logger(ctx).Info("Should insert new ACL for personal folder", resp.Node.ZapPath())
			if e := dao.Add(newAcl); e != nil {
				log.Logger(ctx).Error("-- Could not create recycle_root ACL", zap.Error(e))
			}
		}
		return nil
	}, 8*time.Second)

	return nil
}
