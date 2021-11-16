package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pydio/cells/v4/common/server"
	httpserver "github.com/pydio/cells/v4/common/server/http"
)

// WithHTTP adds a http micro service handler to the current service
func WithHTTP(f func(context.Context, *http.ServeMux) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = httpserver.Default
		o.ServerInit = func() error {
			var srvh *http.Server
			o.Server.(server.Converter).As(&srvh)

			mux, ok := srvh.Handler.(*http.ServeMux)
			if !ok {
				return fmt.Errorf("server is not a mux")
			}
			return f(o.Context, mux)
		}

		// TODO v4 import wrappers for the server
	}
}
