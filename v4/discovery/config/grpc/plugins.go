package grpc

import (
	"context"
	pbconfig "github.com/micro/micro/v3/proto/config"
	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/service"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/plugins"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGrpcNamespace_+common.ServiceConfig),
			service.Context(ctx),
			service.Tag(common.ServiceTagDiscovery),
			service.Description("Main service loading configurations for all other services."),
			// service.WithStorage(config.NewDAO),
			service.WithGRPC(func(srv *grpc.Server) error {
				// Register handler
				pbconfig.RegisterConfigServer(srv, &Handler{})

				return nil
			}),
		)
	})
}
