package server

// ServiceOptions stores all options for a pydio service
type ServerOptions struct {
	// Before and After funcs
	BeforeServe  []func() error
	AfterServe  []func() error
}

// ServerOption is a function to set ServerOptions
type ServerOption func(*ServerOptions)


// BeforeServe executes function before starting the server
func BeforeServe(f func () error) ServerOption {
	return func(o *ServerOptions) {
		o.BeforeServe = append(o.BeforeServe, f)
	}
}

func AfterServe(f func() error) ServerOption {
	return func(o *ServerOptions) {
		o.AfterServe = append(o.AfterServe, f)
	}
}

