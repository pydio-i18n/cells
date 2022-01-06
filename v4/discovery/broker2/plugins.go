package broker

import (
	"context"

	"github.com/pydio/cells/v4/common/broker"

	pb "google.golang.org/genproto/googleapis/pubsub/v1"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
	"google.golang.org/grpc"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGrpcNamespace_+common.ServiceBroker),
			service.Context(ctx),
			service.Tag(common.ServiceTagDiscovery),
			service.Description("Registry"),
			service.WithGRPC(func(ctx context.Context, srv *grpc.Server) error {
				pb.RegisterPublisherServer(srv, NewHandler(broker.Default()))
				pb.RegisterSubscriberServer(srv, NewHandler(broker.Default()))

				return nil
			}),
		)
	})
}
