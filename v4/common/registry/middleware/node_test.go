package middleware

import (
	"context"
	"testing"

	"github.com/pydio/cells/v4/common/registry"
)

func TestNodeRegistry(t *testing.T) {
	r, _ := registry.OpenRegistry(context.Background(), "memory:///")
	r = NewNodeRegistry(r)

	var rn registry.NodeRegistry
	r.As(&rn)

	rn.ListNodes()
}