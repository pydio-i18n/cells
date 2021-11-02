package registry

import "github.com/micro/micro/v3/service/registry"

func toService(ss []*registry.Service) []Service {
	var ret []Service
	for _, s := range ss {
		ret = append(ret, &service{s})
	}
	return ret
}

func toNode(nn []*registry.Node) []Node {
	var ret []Node
	for _, n := range nn {
		ret = append(ret, &node{n})
	}
	return ret
}
