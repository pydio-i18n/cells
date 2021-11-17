/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
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
	"time"

	"github.com/pydio/cells/v4/common/proto/tree"

	"github.com/micro/micro/v3/service/errors"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/micro/broker"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/idm/acl"
)

// Handler definition
type Handler struct {
	idm.UnimplementedACLServiceServer
	tree.UnimplementedNodeProviderStreamerServer
}

// CreateACL in database
func (h *Handler) CreateACL(ctx context.Context, req *idm.CreateACLRequest) (*idm.CreateACLResponse, error) {

	resp := &idm.CreateACLResponse{}
	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	if err := dao.Add(req.ACL); err != nil {
		return nil, err
	}

	resp.ACL = req.ACL
	broker.MustPublish(ctx, common.TopicIdmEvent, &idm.ChangeEvent{
		Type: idm.ChangeEventType_UPDATE,
		Acl:  req.ACL,
	})
	return resp, nil
}

// ExpireACL in database
func (h *Handler) ExpireACL(ctx context.Context, req *idm.ExpireACLRequest) (*idm.ExpireACLResponse, error) {

	resp := &idm.ExpireACLResponse{}

	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	numRows, err := dao.SetExpiry(req.Query, time.Unix(req.Timestamp, 0))
	if err != nil {
		return nil, err
	}

	resp.Rows = numRows

	return resp, nil
}

// DeleteACL from database
func (h *Handler) DeleteACL(ctx context.Context, req *idm.DeleteACLRequest) (*idm.DeleteACLResponse, error) {

	response := &idm.DeleteACLResponse{}
	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	acls := new([]interface{})
	if err := dao.Search(req.Query, acls); err != nil {
		return nil, err
	}

	numRows, err := dao.Del(req.Query)
	response.RowsDeleted = numRows
	if err == nil {
		for _, in := range *acls {
			if val, ok := in.(*idm.ACL); ok {
				broker.MustPublish(ctx, common.TopicIdmEvent, &idm.ChangeEvent{
					Type: idm.ChangeEventType_DELETE,
					Acl:  val,
				})
			}
		}
	}
	return response, err
}

// SearchACL in database
func (h *Handler) SearchACL(request *idm.SearchACLRequest, response idm.ACLService_SearchACLServer) error {

	ctx := response.Context()

	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	acls := new([]interface{})
	if err := dao.Search(request.Query, acls); err != nil {
		return err
	}

	for _, in := range *acls {
		val, ok := in.(*idm.ACL)
		if !ok {
			return errors.InternalServerError(common.ServiceAcl, "Wrong type")
		}
		if e := response.Send(&idm.SearchACLResponse{ACL: val}); e != nil {
			return e
		}
	}

	return nil
}

// StreamACL from database
func (h *Handler) StreamACL(streamer idm.ACLService_StreamACLServer) error {

	ctx := streamer.Context()
	dao := servicecontext.GetDAO(ctx).(acl.DAO)

	for {
		incoming, err := streamer.Recv()
		if incoming == nil || err != nil {
			break
		}

		acls := new([]interface{})
		if err := dao.Search(incoming.Query, acls); err != nil {
			return err
		}

		for _, in := range *acls {
			if val, ok := in.(*idm.ACL); ok {
				if e := streamer.Send(&idm.SearchACLResponse{ACL: val}); e != nil {
					return e
				}
			}
		}

		if e := streamer.Send(nil); e != nil {
			return e
		}
	}

	return nil
}
