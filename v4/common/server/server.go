package server

import (
	"net"
)

type Server interface {
	Serve(net.Listener) error
}

type WrappedServer interface {
	RegisterBeforeServe(func() error)
	BeforeServe() error
	RegisterAfterServe(func () error)
	AfterServe() error
}

type Converter interface {
	As(interface{}) bool
}

type ServerImpl struct {
	opts ServerOptions
}

func (s *ServerImpl) RegisterBeforeServe(f func() error) {
	s.opts.BeforeServe = append(s.opts.BeforeServe, f)
}

func (s *ServerImpl) BeforeServe() error {
	for _, h := range s.opts.BeforeServe {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerImpl) RegisterAfterServe(f func() error) {
	s.opts.AfterServe = append(s.opts.AfterServe, f)
}

func (s *ServerImpl) AfterServe() error {
	for _, h := range s.opts.AfterServe {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}