package main

import (
	"github.com/pydio/cells/v4/cmd"

	_ "github.com/pydio/cells/v4/discovery/config/grpc"
	_ "github.com/pydio/cells/v4/discovery/config/web"

	// Install
	_ "github.com/pydio/cells/v4/discovery/install/rest"

	// Discovery
	_ "github.com/pydio/cells/v4/discovery/health/generic"
	_ "github.com/pydio/cells/v4/discovery/health/grpc"
	_ "github.com/pydio/cells/v4/discovery/health/http"
	_ "github.com/pydio/cells/v4/discovery/registry"

	// Gateways
	_ "github.com/pydio/cells/v4/gateway/proxy"
	// Test Minio Starts (as object, or as gateway)
	//_ "github.com/pydio/cells/v4/data/source/objects/grpc"
	_ "github.com/pydio/cells/v4/gateway/data"
)

func main() {
	cmd.Execute()
}
