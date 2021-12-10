package server

import (
	"context"
	"fmt"
	"os"

	servercontext "github.com/pydio/cells/v4/common/server/context"
)

type Server interface {
	Serve() error
	Stop() error
	Address() []string
	Name() string
	Endpoints() []string
	Metadata() map[string]string
	As(interface{}) bool
}

type WrappedServer interface {
	RegisterBeforeServe(func() error)
	BeforeServe() error
	RegisterAfterServe(func() error)
	AfterServe() error
	RegisterBeforeStop(func() error)
	BeforeStop() error
	RegisterAfterStop(func() error)
	AfterStop() error
}

type server struct {
	s    Server
	opts ServerOptions
}

func NewServer(ctx context.Context, s Server) Server {
	reg := servercontext.GetRegistry(ctx)

	srv := &server{
		s: s,
		opts: ServerOptions{
			Context: ctx,
		},
	}

	reg.Register(srv)

	return srv
}

func (s *server) Serve() error {
	if err := s.BeforeServe(); err != nil {
		return err
	}

	if err := s.s.Serve(); err != nil {
		return err
	}

	if err := s.AfterServe(); err != nil {
		return err
	}

	return nil
}

func (s *server) Stop() error {
	if err := s.BeforeStop(); err != nil {
		return err
	}

	if err := s.s.Stop(); err != nil {
		return err
	}

	if err := s.AfterStop(); err != nil {
		return err
	}

	return nil
}

func (s *server) Address() []string {
	return s.s.Address()
}

func (s *server) Name() string {
	return s.s.Name()
}

func (s *server) Endpoints() []string {
	return s.s.Endpoints()
}

func (s *server) Metadata() map[string]string {
	meta := s.s.Metadata()
	meta["pid"] = fmt.Sprintf("%d", os.Getpid())

	return meta
}

func (s *server) RegisterBeforeServe(f func() error) {
	s.opts.BeforeServe = append(s.opts.BeforeServe, f)
}

func (s *server) BeforeServe() error {
	for _, h := range s.opts.BeforeServe {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) RegisterAfterServe(f func() error) {
	s.opts.AfterServe = append(s.opts.AfterServe, f)
}

func (s *server) AfterServe() error {
	for _, h := range s.opts.AfterServe {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) RegisterBeforeStop(f func() error) {
	s.opts.BeforeStop = append(s.opts.BeforeStop, f)
}

func (s *server) BeforeStop() error {
	for _, h := range s.opts.BeforeStop {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) RegisterAfterStop(f func() error) {
	s.opts.AfterStop = append(s.opts.AfterStop, f)
}

func (s *server) AfterStop() error {
	for _, h := range s.opts.AfterStop {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) As(i interface{}) bool {
	if v, ok := i.(*Server); ok {
		*v = s
		return true
	}
	return s.s.As(i)
}
