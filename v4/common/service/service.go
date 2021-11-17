package service

import (
	"fmt"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/server"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
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

type Service interface{
	Init() error
}

func NewService(opts ...ServiceOption) Service {
	s := &service{
		opts: newOptions(append(mandatoryOptions, opts...)...),
	}

	name := s.opts.Name
	tags := s.opts.Tags

	if !runtime.IsRequired(name) {
		return nil
	}

	if s.opts.Fork && !runtime.IsFork() {
		return nil
	}

	bs, ok := s.opts.Server.(server.WrappedServer)
	if ok {
		bs.RegisterBeforeServe(s.Init)
		bs.RegisterBeforeServe(func() error {
			log.Info("started", zap.String("name", name))
			return nil
		})
		bs.RegisterAfterServe(func() error {
			log.Info("stopped", zap.String("name", name))
			return nil
		})
	}

	registry.Register(name, strings.Join(tags, " "))

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
	r := fmt.Sprintf("grpc://:%d", viper.GetInt("port_registry"))
	//}

	//b := viper.GetString("broker")
	//if b == "memory" {
	b := fmt.Sprintf("grpc://:%d", viper.GetInt("port_broker"))
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