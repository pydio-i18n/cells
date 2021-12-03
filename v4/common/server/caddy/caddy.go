package caddy

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/pprof"
	"reflect"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/proto/install"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/caddy/mux"
)

const (
	caddyfile = `
{
  auto_https disable_redirects
}

{{range .Sites}}
{{$SiteWebRoot := .WebRoot}}
{{range .Binds}}{{.}} {{end}} {
	root * "{{if $SiteWebRoot}}{{$SiteWebRoot}}{{else}}{{$.WebRoot}}{{end}}"
	file_server

	route /* {
		mux
		request_header Host {host}
		request_header X-Real-IP {remote}
	}
	
	{{if .TLS}}tls {{.TLS}}{{end}}
	{{if .TLSCert}}tls "{{.TLSCert}}" "{{.TLSKey}}"{{end}}
}
{{end}}
	 `
)

type Server struct {
	ctx context.Context

	*http.ServeMux
	*server.ServerImpl
	Confs []byte
}

type SiteConf struct {
	*install.ProxyConfig
	// Parsed values from proto oneOf
	TLS     string
	TLSCert string
	TLSKey  string
	// Parsed External host if any
	ExternalHost string
	// Custom Root for this site
	WebRoot string
}

func New(ctx context.Context, dir string) (server.Server, error) {
	srvMUX := http.NewServeMux()
	srvMUX.HandleFunc("/debug/pprof/", pprof.Index)
	srvMUX.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	srvMUX.HandleFunc("/debug/pprof/profile", pprof.Profile)
	srvMUX.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	srvMUX.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.RegisterServerMux(ctx, srvMUX)

	// Creating temporary caddy file
	sites, err := config.LoadSites()
	if err != nil {
		return nil, err
	}

	caddySites, err := SitesToCaddyConfigs(sites)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("pydiocaddy").Parse(caddyfile)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(buf, struct{
		Sites []SiteConf
		WebRoot string
	} {
		caddySites,
		dir,
	}); err != nil {
		return nil, err
	}

	b := buf.Bytes()

	fmt.Println(string(b))

	// Load config directly from memory
	adapter := caddyconfig.GetAdapter("caddyfile")
	confs, _, err := adapter.Adapt(b, map[string]interface{}{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &Server{
		ctx: ctx,
		ServeMux: srvMUX,
		Confs: confs,
		ServerImpl: &server.ServerImpl{},
	}, nil
}

func (s *Server) Serve(l net.Listener) error {
	if err := s.BeforeServe(); err != nil {
		return err
	}

	if err := caddy.Load(s.Confs, true); err != nil {
		return err
	}

	if err := s.AfterServe(); err != nil {
		return err
	}

	select{
	case <-s.ctx.Done():
	}

	if err := s.BeforeStop(); err != nil {
		return err
	}

	if err := caddy.Stop(); err != nil {
		return err
	}

	if err := s.AfterStop(); err != nil {
		return err
	}

	return nil
}

func (s *Server) Address() []string{
	return []string{}
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