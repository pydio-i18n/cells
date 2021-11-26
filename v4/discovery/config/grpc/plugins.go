package grpc

import (
	"context"
	pb "github.com/pydio/cells/v4/common/proto/config"
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
			service.WithGRPC(func(c context.Context, srv *grpc.Server) error {
				// Register handler
				pb.RegisterConfigServer(srv, &Handler{})


				return nil
			}),
		)
	})
}
