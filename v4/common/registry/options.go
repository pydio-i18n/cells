package registry

import (
	"context"

	pb "github.com/pydio/cells/v4/common/proto/registry"
)

type Option func(*Options) error

type Options struct{
	Context context.Context
	Name string
	Type pb.ItemType
}

func WithName(n string) Option {
	return func(o *Options) error {
		o.Name = n
		return nil
	}
}

func WithType(t pb.ItemType) Option {
	return func (o *Options) error {
		o.Type = t
		return nil
	}
}
