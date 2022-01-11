package grpc

import (
	"time"

	"github.com/pydio/cells/v4/common/registry"
	"google.golang.org/grpc"
)

type Option func(*Options)

type Options struct {
	ClientConn  grpc.ClientConnInterface
	Registry    registry.Registry
	CallTimeout time.Duration
	DialOptions []grpc.DialOption
}

func WithClientConn(c grpc.ClientConnInterface) Option {
	return func(o *Options) {
		o.ClientConn = c
	}
}

func WithRegistry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func WithCallTimeout(c time.Duration) Option {
	return func(o *Options) {
		o.CallTimeout = c
	}
}

func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(o *Options) {
		o.DialOptions = append(o.DialOptions, opts...)
	}
}
