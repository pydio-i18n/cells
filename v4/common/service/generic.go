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
		o.ServerInit = func() error {
			var srvg *generic.Server

			o.Server.(server.Converter).As(&srvg)

			return f(o.Context, srvg)
		}

		//		ctx := o.Context

		//		srv, ok := ctx.Value("genericServerKey").(server.Server)
		//		if !ok {
		// log.Println("Context does not contain server key")
		//	return
		//}

		// srv.Handle(f)

		// TODO v4 import wrappers for the server
	}
}
