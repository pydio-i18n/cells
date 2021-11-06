package service

import (
	"net/http"
)

// WithHTTP adds a http micro service handler to the current service
func WithHTTP(f func(*http.ServeMux)) ServiceOption {
	return func(o *ServiceOptions) {
		ctx := o.Context

		mux, ok := ctx.Value("httpServerKey").(*http.ServeMux)
		if !ok {
			// log.Println("Context does not contain server key")
			return
		}

		f(mux)

		// TODO v4 import wrappers for the server
	}
}
