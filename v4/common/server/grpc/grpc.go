package grpc

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	ctx context.Context
	net.Listener
	*grpc.Server
	*server.ServerImpl
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

	return &Server{
		ctx: ctx,
		Server: s,
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

	var err error
	select {
	case err = <-errCh:
		if err != nil {
			return err
		}
	case <-s.ctx.Done():
	}

	if err := s.BeforeStop(); err != nil {
		return err
	}

	s.Server.Stop()

	if err := s.AfterStop(); err != nil {
		return err
	}

	return err
}

func (s *Server) Id() string {
	return "test"
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
