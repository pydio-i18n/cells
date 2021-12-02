package http

import (
	"context"
	"github.com/pydio/cells/v4/common/server"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"
)

type Server struct {
	net.Listener
	*http.ServeMux
	*http.Server
	*server.ServerImpl
}

func New() server.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{}
	srv.Handler = mux

	return &Server{
		ServeMux: mux,
		Server:     srv,
		ServerImpl: &server.ServerImpl{},
	}
}

func (s *Server) Serve(l net.Listener) error {
	lis, err := net.Listen("tcp", viper.GetString("http.address"))
	if err != nil {
		return err
	}
	defer lis.Close()

	s.Listener = lis

	if err := s.BeforeServe(); err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		if err := s.Server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	if err := s.AfterServe(); err != nil {
		return err
	}

	err = <-errCh

	if err := s.BeforeStop(); err != nil {
		errCh <- err
	}

	// todo v4 - probably pass it the context initially
	s.Server.Shutdown(context.TODO())

	if err := s.AfterStop(); err != nil {
		errCh <- err
	}

	return err
}

func (s *Server) Address() []string{
	if s.Listener == nil {
		return []string{}
	}
	return []string{s.Listener.Addr().String()}
}

func (s *Server) Endpoints() []string {
	var endpoints []string
	for _, k := range reflect.ValueOf(s.ServeMux).Elem().FieldByName("m").MapKeys() {
		endpoints = append(endpoints, k.String())
	}

	return endpoints
}

func (s *Server) Id() string {
	return "testhttp"
}

func (s *Server) Metadata() map[string]string {
	return map[string]string{}
}

func (s *Server) As(i interface{}) bool {
	p, ok := i.(**http.ServeMux)
	if !ok {
		return false
	}

	*p = s.ServeMux
	return true
}