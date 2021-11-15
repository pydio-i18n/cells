package http

import (
	"github.com/pydio/cells/v4/common/server"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
	*server.ServerImpl
}

var (
	Default = New()
)

func Register(s *Server) {
	Default = s
}

func New() *Server {
	mux := http.NewServeMux()
	srv := &http.Server{}
	srv.Handler = mux

	return &Server{
		Server: srv,
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
	p, ok := i.(**http.Server)
	if !ok {
		return false
	}

	*p = s.Server
	return true
}
