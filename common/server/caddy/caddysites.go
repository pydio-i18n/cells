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
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pydio/caddyvault"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/config/routing"
	"github.com/pydio/cells/v5/common/crypto/providers"
	"github.com/pydio/cells/v5/common/crypto/storage"
	"github.com/pydio/cells/v5/common/proto/install"
	"github.com/pydio/cells/v5/common/runtime"
	"github.com/pydio/cells/v5/common/telemetry/metrics"
)

type ActiveSite struct {
	*routing.ActiveProxy
	// Final resolve with mux or reverse proxy
	MuxMode bool
	// LogFile for this site
	Log string
	// LogLevel for this site
	LogLevel string
}

// ResolveSites generates a Caddyfile using routing.ResolveProxy with custom resolvers
func ResolveSites(ctx context.Context, resolver routing.UpstreamsResolver, external bool) ([]byte, []string, error) {
	// Creating temporary caddy file
	sites, err := routing.LoadSites(ctx)
	if err != nil {
		return nil, nil, err
	}

	caddySites, err := sitesToCaddySites(sites, resolver)
	if err != nil {
		return nil, nil, err
	}

	tplData := TplData{
		Sites:             caddySites,
		EnableMetrics:     metrics.HasProviders(),
		DisableAdmin:      !external,
		RedirectLogWriter: !external,
		MuxMode:           resolver == nil,
		CorsAllowAll:      os.Getenv("CELLS_WEB_CORS_ALLOW_ALL") == "true",
	}

	k, e := storage.OpenStore(ctx, runtime.CertsStoreURL())
	if e != nil {
		return nil, nil, e
	}
	// Special treatment for vault : append info to caddy
	if vs, ok := k.(*caddyvault.VaultStorage); ok {
		tplData.Storage = `vault {
  address "` + vs.API + `"
  token ` + vs.Token + `
  prefix ` + vs.Prefix + `
}`
	}

	caddyFile, err := FromTemplate(ctx, tplData)
	if err != nil {
		fmt.Println("error eval template", err)
		return nil, nil, err
	}

	var addresses []string
	for _, site := range caddySites {
		for _, bind := range site.GetBinds() {
			//s.addresses = append(s.addresses, bind)

			bind = strings.TrimPrefix(bind, "http://")
			bind = strings.TrimPrefix(bind, "https://")

			host, port, err := net.SplitHostPort(bind)
			if err != nil {
				continue
			}
			ip := net.ParseIP(host)
			if ip == nil || ip.IsUnspecified() {
				addresses = append(addresses, net.JoinHostPort(runtime.DefaultAdvertiseAddress(), port))
			} else {
				addresses = append(addresses, bind)
			}
		}
	}
	return caddyFile, addresses, nil
}

// sitesToCaddySites computes all SiteConf from all *install.ProxyConfig by analyzing
// TLSConfig, ReverseProxyURL and Maintenance fields values
func sitesToCaddySites(sites []*install.ProxyConfig, upstreamResolver routing.UpstreamsResolver) (caddySites []*ActiveSite, er error) {

	rewriteResolver := func(cr *routing.ActiveRoute, route routing.Route, rule *install.Rule) {
		if rule.Action == "Forbidden" {
			cr.Path = strings.TrimSuffix(cr.Path, "/") + "/*"
			cr.RewriteRules = append(cr.RewriteRules, "respond 403")
			return
		}
		// Transform []*install.HeaderMods to []string for CaddyFile
		var stringMods []any
		stringMods = append(stringMods, "request_header X-Real-Ip {http.request.remote}")
		stringMods = append(stringMods, "request_header X-Forwarded-Proto {http.request.scheme}")
		for _, m := range cr.HeaderMods {
			mod := m.(*install.HeaderMod)
			if mod.Action == install.HeaderModAction_REMOVE {
				if mod.ApplyTo == install.HeaderModApplyTo_REQUEST {
					stringMods = append(stringMods, "request_header -"+mod.Key)
				} else {
					stringMods = append(stringMods, "header -"+mod.Key)
				}
			} else {
				appendPlus := ""
				if mod.Action == install.HeaderModAction_ADD_IF_ABSENT || mod.Action == install.HeaderModAction_APPEND_IF_EXISTS_OR_ADD {
					appendPlus = "+"
				}
				if mod.ApplyTo == install.HeaderModApplyTo_REQUEST {
					stringMods = append(stringMods, "request_header "+appendPlus+mod.Key+" "+mod.Value)
				} else {
					stringMods = append(stringMods, "header "+appendPlus+mod.Key+" "+mod.Value)
				}
			}
		}
		cr.HeaderMods = stringMods
		if rule.Action == "Rewrite" {
			inputURI := rule.Value
			realTarget := route.GetURI()
			if realTarget == "/" {
				cr.Path = inputURI + "*"
				cr.RewriteRules = append(cr.RewriteRules, fmt.Sprintf("redir %s %s/", inputURI, inputURI))
				cr.RewriteRules = append(cr.RewriteRules, fmt.Sprintf("uri %s* strip_prefix %s", inputURI, inputURI))
			} else {
				cr.Path = inputURI + "/*"
				cr.RewriteRules = append(cr.RewriteRules, fmt.Sprintf("uri %s/* replace %s/ %s/ 1", inputURI, inputURI, realTarget))
			}
		}
	}

	tlsResolver := func(site *routing.ActiveProxy) error {
		if site.TLSConfig == nil {
			for i, b := range site.Binds {
				site.Binds[i] = "http://" + strings.Replace(b, "0.0.0.0", "", 1)
			}
		} else {
			for i, b := range site.Binds {
				site.Binds[i] = strings.Replace(b, "0.0.0.0", "", 1)
			}
			switch v := site.TLSConfig.(type) {
			case *install.ProxyConfig_Certificate, *install.ProxyConfig_SelfSigned:
				certFile, keyFile, err := providers.LoadCertificates(site.ProxyConfig, runtime.CertsStoreURL())
				if err != nil {
					return err
				}
				site.TLS = fmt.Sprintf(`"%s" "%s"`, certFile, keyFile)
			case *install.ProxyConfig_LetsEncrypt:
				caUrl := common.DefaultCaUrl
				if v.LetsEncrypt.StagingCA {
					caUrl = common.DefaultCaStagingUrl
				}
				site.TLS = v.LetsEncrypt.Email + ` {
				ca ` + caUrl + `
			}`
			}
		}
		return nil
	}

	logLevel := runtime.LogLevel()
	var caddyLogFile, caddyLogLevel string
	if logLevel != "warn" {
		if logLevel == "debug" {
			caddyLogFile = filepath.Join(runtime.ApplicationWorkingDir(runtime.ApplicationDirLogs), "caddy_access.log")
			caddyLogLevel = "INFO"
		} else {
			caddyLogFile = filepath.Join(runtime.ApplicationWorkingDir(runtime.ApplicationDirLogs), "caddy_errors.log")
			caddyLogLevel = "ERROR"
		}
	}

	for _, site := range sites {
		activeProxy, err := routing.ResolveProxy(site, tlsResolver, rewriteResolver, upstreamResolver)
		if err != nil {
			return nil, err
		}
		cs := &ActiveSite{
			ActiveProxy: activeProxy,
			MuxMode:     upstreamResolver == nil,
			Log:         caddyLogFile,
			LogLevel:    caddyLogLevel,
		}
		caddySites = append(caddySites, cs)
	}
	return caddySites, nil
}
