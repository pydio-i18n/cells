package health

import (
	"context"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/registry"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		srv, ok := ctx.Value("grpcServerKey").(*grpc.Server)
		if !ok {
			log.Println("Context does not contain server key")
			return
		}

		registry.Register("health", "discovery")
		grpc_health_v1.RegisterHealthServer(srv, &Handler{})
	})
}
