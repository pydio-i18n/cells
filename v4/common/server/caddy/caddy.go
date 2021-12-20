package caddy

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/http/pprof"
	"reflect"

	caddy "github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/proto/install"
	"github.com/pydio/cells/v4/common/server"
	"github.com/pydio/cells/v4/common/server/caddy/mux"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

const (
	caddyfile = `
{
  auto_https disable_redirects
}

{{range .Sites}}
{{$SiteWebRoot := .WebRoot}}
{{$ExternalHost := .ExternalHost}}
{{range .Binds}}{{.}} {{end}} {
	root * "{{if $SiteWebRoot}}{{$SiteWebRoot}}{{else}}{{$.WebRoot}}{{end}}"

	@grpc-content {
		header Content-type *application/grpc*
	}
	@list_buckets {
		path / /probe-bucket-sign*
		header Authorization *AWS4-HMAC-SHA256*
	}

	route /* {
		# request_header Host {{if $ExternalHost}}{{$ExternalHost}}{{else}}{host}{{end}}
		request_header X-Real-IP {remote}

		# Special rewrite for grpc requests (always sent on root path)
		rewrite @grpc-content /grpc{path}

		# Special rewrite for s3 list buckets (always sent on root path)
		# rewrite @list_buckets /io{path}

		# Apply mux
		mux

		# If mux did not find endpoint, redirect all to root and re-apply mux
		rewrite /* /
		mux
	}


	{{if .TLS}}tls {{.TLS}}{{end}}
	{{if .TLSCert}}tls "{{.TLSCert}}" "{{.TLSKey}}"{{end}}
}
{{end}}
	 `
)

type Server struct {
	name string
	*http.ServeMux
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
	if err := tmpl.Execute(buf, struct {
		Sites   []SiteConf
		WebRoot string
	}{
		caddySites,
		dir,
	}); err != nil {
		return nil, err
	}

	b := buf.Bytes()

	// Load config directly from memory
	adapter := caddyconfig.GetAdapter("caddyfile")
	confs, _, err := adapter.Adapt(b, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	return server.NewServer(ctx, &Server{
		name:     "caddy-" + uuid.New(),
		ServeMux: srvMUX,
		Confs:    confs,
	}), nil
}

func (s *Server) Serve() error {
	return caddy.Load(s.Confs, true)
}

func (s *Server) Stop() error {
	return caddy.Stop()
}

func (s *Server) Address() []string {
	return []string{}
}

func (s *Server) Endpoints() []string {
	var endpoints []string
	for _, k := range reflect.ValueOf(s.ServeMux).Elem().FieldByName("m").MapKeys() {
		endpoints = append(endpoints, k.String())
	}

	return endpoints
}

func (s *Server) Name() string {
	return s.name
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
