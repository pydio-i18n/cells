package http

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"

	"github.com/pydio/cells/v4/common/server/middleware"

	"github.com/spf13/viper"

	"github.com/pydio/cells/v4/common/server"
	servercontext "github.com/pydio/cells/v4/common/server/context"
	"github.com/pydio/cells/v4/common/server/http/registrymux"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

type Server struct {
	id     string
	name   string
	cancel context.CancelFunc
	net.Listener
	*http.ServeMux
	*http.Server
}

func New(ctx context.Context) server.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{}
	srv.Handler = registrymux.NewMiddleware(servercontext.GetRegistry(ctx), mux)
	srv.Handler = ContextMiddlewareHandler(middleware.ClientConnIncomingContext(ctx), srv.Handler)

	ctx, cancel := context.WithCancel(ctx)

	return server.NewServer(ctx, &Server{
		id:       "http-" + uuid.New(),
		name:     "http",
		cancel:   cancel,
		ServeMux: mux,
		Server:   srv,
	})
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", viper.GetString("http.address"))
	if err != nil {
		return err
	}

	s.Listener = lis

	go func() {
		defer s.cancel()

		if err := s.Server.Serve(lis); err != nil {
			// TODO v4 log or summat
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	// Return initial context ?
	return s.Server.Shutdown(context.TODO())
}

func (s *Server) Address() []string {
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

func (s *Server) ID() string {
	return s.id
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) Type() server.ServerType {
	return server.ServerType_HTTP
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
