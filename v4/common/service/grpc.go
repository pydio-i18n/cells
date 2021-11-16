package service

import (
	"github.com/pydio/cells/v4/common/server"
	"google.golang.org/grpc"

	grpcserver "github.com/pydio/cells/v4/common/server/grpc"
)

// WithGRPC adds a service handler to the current service
func WithGRPC(f func(*grpc.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = grpcserver.Default
		o.ServerInit = func() error {
			var srvg *grpc.Server
			o.Server.(server.Converter).As(&srvg)
			return f(srvg)
		}

		// TODO v4 import wrappers for the server
	}
}
