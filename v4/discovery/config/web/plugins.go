package rest

import (
	"context"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceRestNamespace_+common.ServiceConfig),
			service.Context(ctx),
			service.Tag(common.ServiceTagDiscovery),
			service.Description("Configuration"),
			service.Dependency(common.ServiceGrpcNamespace_+common.ServiceConfig, []string{}),
			service.WithWeb(func() service.WebHandler {
				return new(Handler)
			}),
		)
	})
}