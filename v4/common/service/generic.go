package service

import (
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/generic"
)

// WithGeneric adds a http micro service handler to the current service
func WithGeneric(f func(*generic.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = generic.Default
		o.ServerInit = func() error {
			var srvg *generic.Server

			o.Server.(server.Converter).As(&srvg)

			return f(srvg)
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
