package main

import (
	"github.com/pydio/cells/v4/cmd"

	_ "github.com/pydio/cells/v4/discovery/config/grpc"
	_ "github.com/pydio/cells/v4/discovery/config/web"

	_ "github.com/pydio/cells/v4/discovery/registry"
	_ "github.com/pydio/cells/v4/discovery/health"
)

func main() {
	cmd.Execute()
}