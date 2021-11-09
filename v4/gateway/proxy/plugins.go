/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
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

// Package proxy loads a Caddy service to provide a unique access to all services and serve the Javascript frontend.
package proxy

import (
	"context"
	"fmt"
	"github.com/pydio/cells/v4/common/service/generic"

	"go.uber.org/zap"

	caddy "github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/service"
)

var (
	cfile = `
:8181 {
	route /* {
		reverse_proxy :8002 {
			fail_duration 20s
			header_up Host {host}
			header_up X-Real-IP {remote}
		}
	}

	route /a* {
		uri strip_prefix /a
        reverse_proxy pydio.gateway.rest {
			fail_duration 20s
			header_up Host {host}
			header_up X-Real-IP {remote}
		}
 	}
	reverse_proxy /oidc* pydio.web.oauth {
		transport http {
			tls_insecure_skip_verify
		}
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}	
	}
	reverse_proxy /io*   pydio.gateway.data {
		header_up Host {host}
		header_up X-Real-IP {remote}
		header_down Content-Security-Policy "script-src 'none'"
		header_down X-Content-Security-Policy "sandbox"
	}
	reverse_proxy /data* pydio.gateway.data {
		header_up Host {host}
		header_up X-Real-IP {remote}
		header_down Content-Security-Policy "script-src 'none'"
		header_down X-Content-Security-Policy "sandbox"
	}
	route /buckets* {
		uri strip_prefix /buckets
		reverse_proxy /buckets pydio.gateway.data {
			header_up Host {host}
			header_up X-Real-IP {remote}
			header_down Content-Security-Policy "script-src 'none'"
			header_down X-Content-Security-Policy "sandbox"
		}
	}
	
	route /ws* {
		uri strip_prefix /ws
		reverse_proxy pydio.gateway.websocket {
			fail_duration 20s
			header_up Host {host}
			header_up X-Real-IP {remote}
		}
	}
	reverse_proxy /dav* pydio.gateway.dav {
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}
		header_down Content-Security-Policy "script-src 'none'"
		header_down X-Content-Security-Policy "sandbox"
	}
	reverse_proxy /plug/* pydio.web.statics {
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}
		header_down Cache-Control "public, max-age=31536000"
	}
	reverse_proxy /public/* pydio.web.statics {
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}
	}
	route /public/plug/* {
		uri strip_prefix /public
		reverse_proxy pydio.web.statics {
			fail_duration 20s
			header_up Host {host}
			header_up X-Real-IP {remote}
			header_down Cache-Control "public, max-age=31536000"
		}
	}
	reverse_proxy /user/reset-password/* pydio.web.statics {
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}
	}
	reverse_proxy /robots.txt pydio.web.statics {
		fail_duration 20s
		header_up Host {host}
		header_up X-Real-IP {remote}
		
	}
	route /login {
		uri replace /login /gui
		reverse_proxy pydio.web.statics {
			fail_duration 20s
			header_up Host {host}
			header_up X-Real-IP {remote}		
		}
	}
	route /grpc* {
		uri strip_prefix /grpc
		reverse_proxy pydio.gateway.grpc {
			transport http {
				tls_insecure_skip_verify
			}
			fail_duration 20s
		}
	}
	
	reverse_proxy /wopi/ pydio.gateway.wopi {
		header_up Host {host}
		header_up X-Real-IP {remote}
	}
	@grpc-content {
		header Content-type *application/grpc*
	}
	rewrite @grpc-content /grpc/{path}
	@root_standard {
		path /
		not header Content-Type *application/grpc* 
		not header Authorization *AWS4-HMAC-SHA256* 
	}
	@list_buckets {
		path / /probe-bucket-sign*
		header Authorization *AWS4-HMAC-SHA256*
	}
	@uri_standard {
		not path /login /a/* /oidc/* /io/* /data/* /buckets/* /ws/* /plug/* /dav/* /public/* /user/reset-password/* /robots.txt
		not file 
	}
	rewrite @list_buckets /buckets{path}
	redir @root_standard /login 302
	rewrite @uri_standard /login

}
`
)

func init() {
	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGatewayProxy),
			service.Context(ctx),
			service.Tag(common.ServiceTagGateway),
			service.Description("Main HTTP proxy for exposing a unique address to the world"),
			// service.Unique(true),
			service.WithGeneric(func(srv generic.Server) error {
				// Load config directly from memory
				adapter := caddyconfig.GetAdapter("caddyfile")
				confs, warns, err := adapter.Adapt([]byte(cfile), map[string]interface{}{})
				if err != nil {
					fmt.Println(err)
					return err
				}
				for _, w := range warns {
					log.Info("[WARN]", zap.String("", w.String()))
				}
				caddy.TrapSignals()
				fmt.Println("loading caddy")
				if e := caddy.Load(confs, true); e != nil {
					return err
				}
				return nil
			}),
		)
	})
}
