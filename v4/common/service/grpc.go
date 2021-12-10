package service

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/config/runtime"
	servicecontext "github.com/pydio/cells/v4/common/service/context"

)

// WithGRPC adds a service handler to the current service
func WithGRPC(f func(context.Context, *grpc.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		// Making sure the runtime is correct
		if o.Fork && !runtime.IsFork() {
			return
		}

		o.Server = servicecontext.GetServer(o.Context, "grpc")
		o.serverStart = func() error {
			var srvg *grpc.Server
			o.Server.As(&srvg)

			return f(o.Context, srvg)
		}

		// TODO v4 import wrappers for the server
	}
}

func WithGRPCStop(f func(context.Context, *grpc.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.serverStop = func() error {
			var srvg *grpc.Server
			o.Server.As(&srvg)
			return f(o.Context, srvg)
		}
	}
}