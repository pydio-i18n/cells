package grpc

import (
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc"
	"net"
)


type Server struct {
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
				// servicecontext.MetricsUnaryServerInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				servicecontext.SpanStreamServerInterceptor(),
				// servicecontext.MetricsStreamServerInterceptor(),
			),
		),
		ServerImpl: &server.ServerImpl{},
	}
}

func (s *Server) Serve(l net.Listener) error {
	if err := s.BeforeServe(); err != nil {
		return err
	}

	if err := s.Server.Serve(l); err != nil {
		return err
	}

	if err := s.AfterServe(); err != nil {
		return err
	}

	return nil
}

func (s *Server) As(i interface{}) bool {
	p, ok := i.(**grpc.Server)
	if !ok {
		return false
	}

	*p = s.Server
	return true
}
