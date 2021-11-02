package registry

import (
	"github.com/micro/micro/v3/service/registry"
	"strings"
)

type Service interface{
	Name() string
	Nodes() []Node
	Tags() []string

	IsGeneric() bool
	IsGRPC() bool
	IsREST() bool
}

type service struct {
	s *registry.Service
}

func (s *service) Name() string {
	return s.s.Name
}
func (s *service) Nodes() []Node{
	return toNode(s.s.Nodes)
}
func (s *service) Tags() []string {
	return strings.Split(s.s.Metadata["tags"], ",")
}
func (s *service) IsGeneric() bool {
	return false
}
func (s *service) IsGRPC() bool {
	return true
}
func (s *service) IsREST() bool {
	return false
}
