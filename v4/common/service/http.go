package service

import (
	"context"
	"fmt"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"net/http"

	"github.com/pydio/cells/v4/common/server"
)

// WithHTTP adds a http micro service handler to the current service
func WithHTTP(f func(context.Context, *http.ServeMux) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.Server = servicecontext.GetServer(o.Context, "http")
		o.serverStart = func() error {
			var mux *http.ServeMux
			if !o.Server.(server.Converter).As(&mux) {
				return fmt.Errorf("server is not a mux")
			}

			return f(o.Context, mux)
		}

		// TODO v4 import wrappers for the server
	}
}

func WithHTTPStop(f func(context.Context, *http.ServeMux) error) ServiceOption {
	return func(o *ServiceOptions) {
		o.serverStop = func() error {
			var mux *http.ServeMux
			o.Server.(server.Converter).As(&mux)
			return f(o.Context, mux)
		}
	}
}
