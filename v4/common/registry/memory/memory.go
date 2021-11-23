package memory

import (
	"context"
	"github.com/micro/micro/v3/service/registry/memory"
	"net/url"

	"github.com/pydio/cells/v4/common/registry"
)

var scheme = "memory"

type URLOpener struct {}

func init() {
	o := &URLOpener{}
	registry.DefaultURLMux().Register(scheme, o)
}

func (o *URLOpener) OpenURL(ctx context.Context, u *url.URL) (registry.Registry, error) {
	return registry.New(memory.NewRegistry()), nil
}



