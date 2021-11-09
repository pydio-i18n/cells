package main

import (
	"github.com/pydio/cells/v4/cmd"

	_ "github.com/pydio/cells/v4/discovery/config/grpc"
	_ "github.com/pydio/cells/v4/discovery/config/web"

	// Install
	_ "github.com/pydio/cells/v4/discovery/install/rest"


	// Discovery
	_ "github.com/pydio/cells/v4/discovery/registry"
	_ "github.com/pydio/cells/v4/discovery/health/grpc"
	_ "github.com/pydio/cells/v4/discovery/health/generic"
	_ "github.com/pydio/cells/v4/discovery/health/http"

	 _ "github.com/pydio/cells/v4/gateway/proxy"
)

func main() {
	cmd.Execute()
}