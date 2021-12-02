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

package service

import (
	"github.com/pydio/cells/v4/common"
	pb "github.com/pydio/cells/v4/common/proto/registry"
	"github.com/pydio/cells/v4/common/registry"
	"strings"
)

type service struct {
	s *pb.Service
}

func (s *service) Name() string {
	return s.s.Name
}

func (s *service) Version() string {
	return s.s.Version
}

func (s *service) Metadata() map[string] string {
	return s.s.Metadata
}

func (s *service) Nodes() []registry.Node {
	var nodes []registry.Node
	for _, n := range s.s.Nodes {
		nodes = append(nodes, &node{n})
	}
	return nodes
}

func (s *service) Tags() []string {
	return strings.Split(s.s.Metadata["tags"], ",")
}

func (s *service) IsGRPC() bool {
	return strings.HasPrefix(s.s.Name, common.ServiceGrpcNamespace_)
}

func (s *service) IsREST() bool {
	return strings.HasPrefix(s.s.Name, common.ServiceRestNamespace_)
}

func (s *service) IsGeneric() bool {
	return !s.IsGRPC() && !s.IsREST()
}


type node struct {
	n *pb.Node
}

func (n *node) Id() string {
	return n.n.Id
}

func (n *node) Address() []string {
	return []string{n.n.Address}
}

func (n *node) Endpoints() []string {
	return n.n.Endpoints
}

func (n *node) Metadata() map[string]string {
	return n.n.Metadata
}

type endpoint struct {
	e *pb.Endpoint
}

func (e *endpoint) Name() string {
	return e.e.Name
}

func (e *endpoint) Metadata() map[string]string {
	return e.e.Metadata
}

func ToProtoService(s registry.Service) *pb.Service {
	if ss, ok := s.(*service); ok {
		return ss.s
	}

	var nodes []*pb.Node

	for _, n := range s.Nodes() {
		nodes = append(nodes, ToProtoNode(n))
	}

	return &pb.Service{
		Name:      s.Name(),
		Version:   s.Version(),
		// Metadata:  s.Metadata(),
		// Endpoints: endpoints,
		Nodes:     nodes,
		Options:   new(pb.Options),
	}
}

func ToProtoNode(n registry.Node) *pb.Node {
	if nn, ok := n.(*node); ok {
		return nn.n
	}

	// TODO v4
	address := ""
	if len(n.Address()) > 0 {
		address = n.Address()[0]
	}

	return &pb.Node{
		Id:      n.Id(),
		Address:   address,
		Endpoints: n.Endpoints(),
		Metadata:  n.Metadata(),
	}
}

func ToService(s *pb.Service) registry.Service {
	return &service{s}
}

func ToNode(n *pb.Node) registry.Node {
	return &node{n}
}