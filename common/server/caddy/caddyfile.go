/*
 * Copyright (c) 2024. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package caddy

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"text/template"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/telemetry/log"
)

const (
	caddytemplate = `
{
  auto_https disable_redirects
{{if .DisableAdmin}}  admin off{{end}}
{{if .Storage}}  storage {{.Storage}}{{end}}
{{if .RedirectLogWriter}}  log{
     output cells
 	 format json
  }{{end}}
}

{{$CorsAllowAll := .CorsAllowAll}}
{{if .CorsAllowAll}}
(cors) {
  @cors_preflight {
	method OPTIONS
	header_regexp acr Access-Control-Request-Method .+
  }
  @cors header Origin {args[0]}

  handle @cors_preflight {
    header {
		Access-Control-Allow-Origin "{args[0]}"
		Access-Control-Allow-Methods "OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE"
		Access-Control-Allow-Headers "*"
		Access-Control-Max-Age "3600"
	}
    respond "" 204
  }

  handle @cors {
    header Access-Control-Allow-Origin "{args[0]}"
	header Vary Origin
    header Access-Control-Expose-Headers "Authorization,ETag"
	header Access-Control-Allow-Credentials "true"
  }
}
{{end}}

{{range .Sites}}
{{$MuxMode := .MuxMode}}
{{$SiteHash := .Hash}}
{{$Maintenance := .Maintenance}}
{{$MaintenanceConditions := .MaintenanceConditions}}
{{range .Binds}}{{.}} {{end}} {

	{{range .Routes}}
	route {{.Path}} {

		{{if $CorsAllowAll}}import cors {header.origin}{{end}}
		{{range .HeaderMods}}{{.}}
		{{end}}

		{{if $Maintenance}}
		# Special redir for maintenance mode
		@rmatcher {
			{{range $MaintenanceConditions}}{{.}}
			{{end}}
			not path /maintenance.html
		}
		request_header X-Maintenance-Redirect "true"
		redir @rmatcher /maintenance.html
		{{end}}		

		{{range .RewriteRules}}{{.}}
		{{end}}
		{{if $MuxMode}}
		# Apply mux
		mux
		{{else}}
		reverse_proxy {{joinUpstreams .Upstreams " "}} {{if $CorsAllowAll}}{
			header_down -Access-Control-Allow-Origin
		}{{end}}
		{{end}}
	}
	{{end}}

	{{if .Log}}
	log {
		output file "{{.Log}}"
		level {{.LogLevel}}
	}
	{{end}}

	{{if .TLS}}tls {{.TLS}}{{end}}
}
{{if .SSLRedirect}}
{{range $k,$v := .Redirects}}
{{$k}} {
	redir {{$v}}
}
{{end}}
{{end}}
{{end}}
	 `
)

type TplData struct {
	Sites             []*ActiveSite
	Storage           string
	MuxMode           bool
	EnableMetrics     bool
	DisableAdmin      bool
	RedirectLogWriter bool
	CorsAllowAll      bool
}

var (
	parsedTpl  *template.Template
	parsedOnce sync.Once
)

func joinUpstreams(uu []any, sep string) string {
	var addr []string
	for _, u := range uu {
		if s, o := u.(string); o {
			addr = append(addr, s)
		} else if ur, o2 := u.(*url.URL); o2 {
			addr = append(addr, ur.String())
		}
	}
	return strings.Join(addr, sep)
}

func FromTemplate(ctx context.Context, tplData TplData) ([]byte, error) {
	var err error
	parsedOnce.Do(func() {
		parsedTpl, err = template.New("pydiocaddy").Funcs(template.FuncMap{"joinUpstreams": joinUpstreams}).Parse(caddytemplate)
	})
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := parsedTpl.Execute(buf, tplData); err != nil {
		return nil, err
	}

	b := buf.Bytes()
	b = caddyfile.Format(b)

	if common.LogLevel == zap.DebugLevel {
		fmt.Println(string(b))
	}

	// Load config directly from memory
	adapter := caddyconfig.GetAdapter("caddyfile")
	confs, ww, err := adapter.Adapt(b, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	for _, w := range ww {
		log.Logger(ctx).Warn(w.String())
	}
	return confs, nil
}
