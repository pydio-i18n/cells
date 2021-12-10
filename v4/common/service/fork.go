package service

func Fork(f bool) ServiceOption {
	return func(o *ServiceOptions) {
		o.Fork = f
	}
}
