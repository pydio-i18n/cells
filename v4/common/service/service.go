package service

import (
	"github.com/pydio/cells/v4/common/config/runtime"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/server"
	"go.uber.org/zap"
	"strings"
)

// Service for the pydio app
type service struct {
	opts *ServiceOptions
}

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