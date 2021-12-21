package service

import (
	"fmt"

	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/server"

	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/spf13/viper"
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
	Init(opts ...ServiceOption)
	Options() *ServiceOptions
	Metadata() map[string]string
	Name() string
	Start() error
	Stop() error
	IsGRPC() bool
	IsREST() bool
	IsGeneric() bool
	As(i interface{}) bool
}

type Stopper func() error

func NewService(opts ...ServiceOption) Service {
	s := &service{
		opts: newOptions(append(mandatoryOptions, opts...)...),
	}

	name := s.opts.Name

	s.opts.Context = servicecontext.WithServiceName(s.opts.Context, name)

	reg := servicecontext.GetRegistry(s.opts.Context)

	reg.Register(s)

	return s
}

func (s *service) Init(opts ...ServiceOption) {
	for _, o := range opts {
		if o == nil {
			continue
		}

		o(s.opts)
	}
}

func (s *service) Options() *ServiceOptions {
	return s.opts
}

func (s *service) Metadata() map[string]string {
	return s.opts.Metadata
}

func (s *service) As(i interface{}) bool {
	if v, ok := i.(*Service); ok {
		*v = s
		return true
	}

	return false
}

func (s *service) Start() error {
	for _, before := range s.opts.BeforeStart {
		if err := before(s.opts.Context); err != nil {
			return err
		}
	}

	if s.opts.serverStart != nil {
		if err := s.opts.serverStart(); err != nil {
			return err
		}
	}

	for _, after := range s.opts.AfterStart {
		if err := after(s.opts.Context); err != nil {
			return err
		}
	}

	//log.Logger(s.opts.Context).Info("update service version")
	//if er := UpdateServiceVersion(s.opts); er != nil {
	//	return er
	//}

	log.Logger(s.opts.Context).Info("started")

	return nil
}

func (s *service) Stop() error {
	for _, before := range s.opts.BeforeStop {
		if err := before(s.opts.Context); err != nil {
			return err
		}
	}

	if s.opts.serverStop != nil {
		if err := s.opts.serverStop(); err != nil {
			return err
		}
	}

	for _, after := range s.opts.AfterStop {
		if err := after(s.opts.Context); err != nil {
			return err
		}
	}

	log.Logger(s.opts.Context).Info("stopped")

	return nil
}

func buildForkStartParams(serviceName string) []string {

	r := fmt.Sprintf("grpc://%s", viper.GetString("grpc.address"))
	b := fmt.Sprintf("grpc://%s", viper.GetString("grpc.address"))

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
	return s.opts.serverType == server.ServerType_GENERIC
}
func (s *service) IsGRPC() bool {
	return s.opts.serverType == server.ServerType_GRPC
}
func (s *service) IsREST() bool {
	return s.opts.serverType == server.ServerType_HTTP
}
