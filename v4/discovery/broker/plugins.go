package config

import (
	"context"
	"log"

	pbbroker "github.com/micro/micro/v3/proto/broker"
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

		registry.Register("broker", "discovery")

		pbbroker.RegisterBrokerServer(srv, &Handler{})
	})
}
