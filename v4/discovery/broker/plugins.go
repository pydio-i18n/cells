package broker

import (
	"context"
	"github.com/pydio/cells/v4/common/plugins"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		/*
			srv, ok := ctx.Value("grpcServerKey").(*grpc.Server)
			if !ok {
				log.Println("Context does not contain server key")
				return
			}

			registry.Register("broker", "discovery")

			pbbroker.RegisterBrokerServer(srv, &Handler{})

		*/
	})
}
