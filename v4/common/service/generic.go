package service

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/generic"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

// WithGeneric adds a http micro service handler to the current service
func WithGeneric(f func(context.Context, *generic.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = servicecontext.GetServer(o.Context, "generic")
		o.serverStart = func() error {
			var srvg *generic.Server

			o.Server.(server.Converter).As(&srvg)

			return f(o.Context, srvg)
		}

		// TODO v4 import wrappers for the server
	}
}

// WithGenericStop adds a http micro service handler to the current service
func WithGenericStop(f func(context.Context, *generic.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.serverStop = func() error {
			var srvg *generic.Server

			o.Server.(server.Converter).As(&srvg)

			return f(o.Context, srvg)
		}
	}
}
