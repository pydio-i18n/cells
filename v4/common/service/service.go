package service

import (
	"fmt"
	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/registry"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

// Service for the pydio app
type service struct {
	opts *ServiceOptions
}

const (
	configSrvKeyFork      = "fork"
	configSrvKeyAutoStart = "autostart"
	configSrvKeyForkDebug = "debugFork"
	configSrvKeyUnique    = "unique"
)

var (
	mandatoryOptions []ServiceOption
)

type Service interface {
	Init() error
}

func NewService(opts ...ServiceOption) Service {
	s := &service{
		opts: newOptions(append(mandatoryOptions, opts...)...),
	}

	name := s.opts.Name

	s.opts.Context = servicecontext.WithServiceName(s.opts.Context, name)

	if !runtime.IsRequired(name) {
		return nil
	}

	if s.opts.Fork && !runtime.IsFork() {
		return nil
	}

	reg := servicecontext.GetRegistry(s.opts.Context)

	bs, ok := s.opts.Server.(server.WrappedServer)
	if ok {
		bs.RegisterBeforeServe(s.Init)
		bs.RegisterAfterServe(func() error {
			// Register service again to update nodes information
			if err := reg.RegisterService(s); err != nil {
				return err
			}

			log.Info("started", zap.String("name", name))
			return nil
		})
	}

	reg.RegisterService(s)

	return s
}

func (s *service) Init() error {
	for _, before := range s.opts.BeforeInit {
		if err := before(s.opts.Context); err != nil {
			return err
		}
	}

	if err := s.opts.ServerInit(); err != nil {
		return err
	}

	for _, after := range s.opts.AfterInit {
		if err := after(s.opts.Context); err != nil {
			return err
		}
	}

	return nil
}

func buildForkStartParams(serviceName string) []string {

	//r := viper.GetString("registry")
	//if r == "memory" {
	r := fmt.Sprintf("grpc://%s", viper.GetString("grpc.address"))
	//}

	//b := viper.GetString("broker")
	//if b == "memory" {
	b := fmt.Sprintf("grpc://%s", viper.GetString("grpc.address"))
	//}

	params := []string{
		"start",
		"--fork",
		"--grpc.address", ":0",
		"--http.address", ":0",
		"--config", viper.GetString("config"),
		"--registry", r,
		"--broker", b,
	}
	if viper.GetBool("enable_metrics") {
		params = append(params, "--enable_metrics")
	}
	if viper.GetBool("enable_pprof") {
		params = append(params, "--enable_pprof")
	}
	if config.Get("services", serviceName, configSrvKeyForkDebug).Bool() /*|| strings.HasPrefix(serviceName, "pydio.grpc.data.")*/ {
		params = append(params, "--log", "debug")
	}
	// Use regexp to specify that we want to start that specific service
	params = append(params, "^"+serviceName+"$")
	bindFlags := config.DefaultBindOverrideToFlags()
	if len(bindFlags) > 0 {
		params = append(params, bindFlags...)
	}
	return params
}

func (s *service) Name() string {
	return s.opts.Name
}
func (s *service) Version() string {
	return s.opts.Version
}
func (s *service) Nodes() []registry.Node {
	if s.opts.Server == nil {
		return []registry.Node{}
	}
	return []registry.Node{s.opts.Server}
}
func (s *service) Tags() []string {
	return s.opts.Tags
}
func (s *service) IsGeneric() bool {
	return !s.IsGRPC() && !s.IsREST()
}
func (s *service) IsGRPC() bool {
	return strings.HasPrefix(s.Name(), common.ServiceGrpcNamespace_)
}
func (s *service) IsREST() bool {
	return strings.HasPrefix(s.Name(), common.ServiceRestNamespace_)
}
