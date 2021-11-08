package generic

import (
	"context"
	"fmt"
	"github.com/pydio/cells/v4/common/service/generic"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceTestNamespace_+common.ServiceHealthCheck),
			service.Context(ctx),
			service.Tag(common.ServiceTagDiscovery),
			service.Description("Service launching a test discovery server."),
			// service.WithStorage(config.NewDAO),
			service.WithGeneric(func(srv generic.Server) error {
				fmt.Println("This is a new handler")

				select {
				case <-ctx.Done():
					fmt.Println("Handler is done")
				}

				return nil
			}),
		)
	})
}
