package registry

import (
	"context"
	"github.com/pydio/cells/v4/common"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/plugins"
	pbregistry "github.com/micro/micro/v3/proto/registry"
	"github.com/pydio/cells/v4/common/service"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGrpcNamespace_ + common.ServiceRegistry),
			service.Context(ctx),
			service.Tag(common.ServiceTagDiscovery),
			service.Description("Registry"),
			service.WithGRPC(func (ctx context.Context, srv *grpc.Server) error {
				pbregistry.RegisterRegistryServer(srv, &Handler{})

				return nil
			}),
		)
	})
}
