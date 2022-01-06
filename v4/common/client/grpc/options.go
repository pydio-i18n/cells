package grpc

import (
	"time"

	"google.golang.org/grpc"
)

type Option func(*Options)

type Options struct {
	Registry    string
	CallTimeout time.Duration
	DialOptions []grpc.DialOption
}

func WithRegistry(r string) Option {
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
