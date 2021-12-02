package service

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc"
)

// WithGRPC adds a service handler to the current service
func WithGRPC(f func(context.Context, *grpc.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = servicecontext.GetServer(o.Context, "grpc")
		o.ServerInit = func() error {
			var srvg *grpc.Server
			o.Server.(server.Converter).As(&srvg)
			return f(o.Context, srvg)
		}

		// TODO v4 import wrappers for the server
	}
}
