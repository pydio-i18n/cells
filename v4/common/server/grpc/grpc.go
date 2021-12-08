package grpc

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	cancel context.CancelFunc
	net.Listener
	*grpc.Server
}

func New(ctx context.Context) server.Server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			servicecontext.SpanUnaryServerInterceptor(),
			servicecontext.MetricsUnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			servicecontext.SpanStreamServerInterceptor(),
			servicecontext.MetricsStreamServerInterceptor(),
		),
	)

	ctx, cancel := context.WithCancel(ctx)

	return server.NewServer(ctx, &Server{
		cancel: cancel,
		Server: s,
	})
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", viper.GetString("grpc.address"))
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
	return "testgrpc"
}

func (s *Server) Metadata() map[string]string {
	return map[string]string{}
}

func (s *Server) Address() []string{
	return []string{s.Listener.Addr().String()}
}

func (s *Server) Endpoints() []string{
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
