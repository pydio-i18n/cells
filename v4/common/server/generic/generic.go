package generic

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	"net"
)

type Server struct {
	*server.ServerImpl

	ctx      context.Context
	handlers []func() error
}

type Handler interface {
	Start() error
	Stop() error
}

func New(ctx context.Context) *Server {
	return &Server{
		ctx: ctx,
		ServerImpl: &server.ServerImpl{},
	}
}

func (s *Server) RegisterHandler(h Handler) {
	s.Handle(h.Start)
	s.RegisterAfterServe(h.Stop)
}

func (s *Server) Handle(h func() error) {
	s.handlers = append(s.handlers, h)
}

func (s *Server) Serve(l net.Listener) error {
	if err := s.BeforeServe(); err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		for _, handler := range s.handlers {
			go handler()
		}


		/* TODO improve that */
		if err := s.BeforeStop(); err != nil {
			errCh <- err
		}



		if err := s.AfterStop(); err != nil {
			errCh <- err
		}
	}()

	if err := s.AfterServe(); err != nil {
		return err
	}

	err := <-errCh

	return err
}

func (s *Server) Id() string {
	return "testgeneric"
}

func (s *Server) Metadata() map[string]string {
	return map[string]string{}
}

func (s *Server) Address() []string {
	return []string{}
}

func (s *Server) Endpoints() []string {
	return []string{}
}

func (s *Server) As(i interface{}) bool {
	p, ok := i.(**Server)
	if !ok {
			return false
		}

	*p = s
	return true
}