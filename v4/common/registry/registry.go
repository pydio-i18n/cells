package registry

import (
	"github.com/micro/micro/v3/service/registry"
	"github.com/micro/micro/v3/service/registry/memory"
)

var reg = memory.NewRegistry()

func Registry() registry.Registry {
	return reg
}

func SetRegistry(r registry.Registry) {
	reg = r
}

func Register(name string, tags string) {
	reg.Register(&registry.Service{
		Name: name,
		Version: "0.0.0",
		Metadata: map[string]string{
			"tags": tags,
		},
		Nodes: []*registry.Node{{
			Id: name + "-0",
			Address: ":8001",
		}},
	})
}

func ListServices() ([]Service, error) {
	ss, err := reg.ListServices()
	if err != nil {
		return nil, err
	}

	return toService(ss), nil
}