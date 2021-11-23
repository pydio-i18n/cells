package grpc

import (
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	net.Listener
	*grpc.Server
	*server.ServerImpl
}

var (
	Default = New()
)

func Register(s server.Server) {
	Default = s
}

func New() server.Server {
	return &Server{
		Server: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				servicecontext.SpanUnaryServerInterceptor(),
				servicecontext.MetricsUnaryServerInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				servicecontext.SpanStreamServerInterceptor(),
				servicecontext.MetricsStreamServerInterceptor(),
			),
		),
		ServerImpl: &server.ServerImpl{},
	}
}

func (s *Server) Serve(l net.Listener) error {
	s.Listener = l

	if err := s.BeforeServe(); err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		if err := s.Server.Serve(l); err != nil {
			errCh <- err
		}
	}()

	if err := s.AfterServe(); err != nil {
		return err
	}

	err := <-errCh

	if err := s.BeforeStop(); err != nil {
		errCh <- err
	}

	s.Server.GracefulStop()

	if err := s.AfterStop(); err != nil {
		errCh <- err
	}

	return err
}

func (s *Server) Address() []string{
	return []string{s.Listener.Addr().String()}
}

func (s *Server) As(i interface{}) bool {
	p, ok := i.(**grpc.Server)
	if !ok {
		return false
	}

	*p = s.Server
	return true
}
