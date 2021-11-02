package service

import (
	"google.golang.org/grpc"
	"log"
)

// WithGRPC adds a micro service handler to the current service
func WithGRPC(f func(*grpc.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		ctx := o.Context

		srv, ok := ctx.Value("grpcServerKey").(*grpc.Server)
		if !ok {
			log.Println("Context does not contain server key")
			return
		}

		f(srv)
	}
}