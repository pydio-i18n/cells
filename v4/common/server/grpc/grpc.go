package grpc

import (
	"context"
	"net"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

type Server struct {
	name   string
	cancel context.CancelFunc
	addr   string
	net.Listener
	*grpc.Server
}

// New creates the generic grpc.Server
func New(ctx context.Context) server.Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			servicecontext.ContextUnaryServerInterceptor(servicecontext.SpanIncomingContext),
			servicecontext.MetricsUnaryServerInterceptor(),
			servicecontext.ContextUnaryServerInterceptor(servicecontext.MetaIncomingContext),
		),
		grpc.ChainStreamInterceptor(
			servicecontext.ContextStreamServerInterceptor(servicecontext.SpanIncomingContext),
			servicecontext.MetricsStreamServerInterceptor(),
			servicecontext.ContextStreamServerInterceptor(servicecontext.MetaIncomingContext),
		),
	)

	ctx, cancel := context.WithCancel(ctx)

	return server.NewServer(ctx, &Server{
		name:   "grpc-" + uuid.New(),
		cancel: cancel,
		addr:   viper.GetString("grpc.address"),
		Server: s,
	})
}

// NewWithServer can pass preset grpc.Server with custom listen address
func NewWithServer(ctx context.Context, s *grpc.Server, listen string) server.Server {
	ctx, cancel := context.WithCancel(ctx)
	return server.NewServer(ctx, &Server{
		name:   "grpc-" + uuid.New(),
		cancel: cancel,
		addr:   listen,
		Server: s,
	})

}

func (s *Server) Serve() error {
	//fmt.Println("Serving Grpc on " + s.addr)
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.Listener = lis

	go func() {
		defer s.cancel()

		if err := s.Server.Serve(lis); err != nil {
			// TODO v4 - log or summat
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.Server.Stop()

	return s.Listener.Close()

}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) Metadata() map[string]string {
	return map[string]string{}
}

func (s *Server) Address() []string {
	if s.Listener == nil {
		return []string{}
	}
	return []string{s.Listener.Addr().String()}
}

func (s *Server) Endpoints() []string {
	var endpoints []string

	info := s.Server.GetServiceInfo()
	for _, i := range info {
		for _, m := range i.Methods {
			endpoints = append(endpoints, m.Name)
		}
	}

	return endpoints
}

func (s *Server) As(i interface{}) bool {
	p, ok := i.(**grpc.Server)
	if !ok {
		return false
	}

	*p = s.Server
	return true
}
