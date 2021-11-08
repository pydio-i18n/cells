package service

import (
	"github.com/pydio/cells/v4/common/service/generic"
)

// WithHTTP adds a http micro service handler to the current service
func WithGeneric(f func(generic.Server) error) ServiceOption {
	return func(o *ServiceOptions) {
		ctx := o.Context

		srv, ok := ctx.Value("genericServerKey").(generic.Server)
		if !ok {
			// log.Println("Context does not contain server key")
			return
		}

		srv.Handle(f)

		// TODO v4 import wrappers for the server
	}
}
