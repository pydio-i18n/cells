package service

import (
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/server/fork"
)

func Fork(f bool) ServiceOption {
	return func(o *ServiceOptions) {
		o.Fork = f

		if o.Fork && runtime.IsFork() {
			return
		}

		o.Server = fork.NewServer(o.Context)
		o.serverStart = func () error {
			var srvf *fork.ForkServer

			o.Server.As(&srvf)

			srvf.RegisterForkParam(o.Name)

			return nil
		}
	}
}
