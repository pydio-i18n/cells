package registry

import "github.com/micro/micro/v3/service/registry"

type Node interface{
	Id() string
	Address() string
	Metadata() map[string]string
}

type node struct {
	n *registry.Node
}

func (n *node) Id() string {
	return n.n.Id
}
func (n *node) Address() string {
	return n.n.Address
}
func (n *node) Metadata() map[string]string {
	return n.n.Metadata
}
