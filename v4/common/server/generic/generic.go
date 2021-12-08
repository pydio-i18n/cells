package generic

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
)

type Server struct {
	cancel      context.CancelFunc
	handlers []func() error
}

type Handler interface {
	Start() error
	Stop() error
}

func New(ctx context.Context) server.Server {
	ctx, cancel := context.WithCancel(ctx)
	return server.NewServer(ctx, &Server{
		cancel: cancel,
	})
}

func (s *Server) RegisterHandler(h Handler) {
	s.Handle(h.Start)
}

func (s *Server) Handle(h func() error) {
	s.handlers = append(s.handlers, h)
}

func (s *Server) Serve() error {
	go func() {
		defer s.cancel()

		for _, handler := range s.handlers {
			go handler()
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	return nil
}

func (s *Server) Name() string {
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