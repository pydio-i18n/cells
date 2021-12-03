package middleware

import (
	"github.com/pydio/cells/v4/common/registry"
	"os"
)

type NodeRegistry struct {
	nodes []registry.Node

	r registry.Registry
}

func NewNodeRegistry(r registry.Registry) registry.Registry {
	return &NodeRegistry{
		r: r,
	}
}

func (r *NodeRegistry) StartService(s string) error {
	return r.r.StartService(s)
}

func (r *NodeRegistry) StopService(s string) error {
	return r.r.StopService(s)
}

func (r *NodeRegistry) RegisterService(service registry.Service) error {
	// First register all nodes
	for _, node := range service.Nodes() {
		if err := r.RegisterNode(node); err != nil {
			return err
		}
	}

	return r.r.RegisterService(service)
}

func (r *NodeRegistry) DeregisterService(service registry.Service) error {
	return r.r.DeregisterService(service)
}

func (r *NodeRegistry) GetService(s string) (registry.Service, error) {
	return r.r.GetService(s)
}

func (r *NodeRegistry) ListServices() ([]registry.Service, error) {
	return r.r.ListServices()
}

func (r *NodeRegistry) WatchServices(option ...registry.WatchOption) (registry.Watcher, error) {
	return r.r.WatchServices(option...)
}

func (r *NodeRegistry) RegisterNode(node registry.Node) error {
	for k, v := range r.nodes {
		if v.Id() == node.Id() {
			r.nodes[k] = node
			return nil
		}
	}

	r.nodes = append(r.nodes, node)
	return nil
}

func (r *NodeRegistry) DeregisterNode(node registry.Node) error {
	for k, v := range r.nodes {
		if node.Id() == v.Id() {
			r.nodes = append(r.nodes[:k], r.nodes[k+1:]...)
		}
	}
	return nil
}

func (r *NodeRegistry) GetNode(s string) (registry.Node, error) {
	for _, v := range r.nodes {
		if s == v.Id() {
			return v, nil
		}
	}

	return nil, os.ErrNotExist
}

func (r *NodeRegistry) ListNodes() ([]registry.Node, error) {
	return r.nodes, nil
}

func (r *NodeRegistry) As(i interface{}) bool {
	if v, ok := i.(*registry.NodeRegistry); ok {
		*v = r
		return true
	}

	return r.r.As(i)
}