/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
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
	"html/template"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/registry"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/caddy/hooks"
	_ "github.com/pydio/cells/v4/common/caddy/proxy"
	"github.com/pydio/cells/v4/common/log"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

const (
	caddyRestartDebounce = 5 * time.Second
)

var (
	mainCaddy = &Caddy{}
	FuncMap   = template.FuncMap{
		"urls": internalURLFromServices,
	}
	restartRequired    bool
	gatewayCtx         = servicecontext.WithServiceName(context.Background(), common.ServiceGatewayProxy)
	LastKnownCaddyFile string
	dirOnce            *sync.Once
)

func init() {
	// TODO v4 verify this
	// caddy.AppName = common.PackageLabel
	// caddy.AppVersion = common.Version().String()
	// httpserver.GracefulTimeout = 30 * time.Second
	dirOnce = &sync.Once{}

	go watchRestart()
	go watchStop()
}

func watchRestart() {
	for {
		select {
		case <-hooks.RestartChan:
			log.Logger(context.Background()).Debug("Received Proxy Restart Event")
			restartRequired = true
		case <-time.After(caddyRestartDebounce):
			if restartRequired {
				log.Logger(context.Background()).Debug("Restarting Proxy Now")
				restartRequired = false
				restart()
			}
		}
	}
}

func watchStop() {
	for range hooks.StopChan {
		Stop()
	}
}

// Caddy contains the templates and functions for building a dynamic caddyfile
type Caddy struct {
	caddyfile     string
	caddytemplate *template.Template
	player        hooks.TemplateFunc
	instance      *caddy.Instance
}

// Enable the caddy builder
func Enable(caddyfile string, player hooks.TemplateFunc) {
	dirOnce.Do(func() {
		httpserver.RegisterDevDirective("pydioproxy", "proxy")
	})
	caddytemplate, err := template.New("pydiocaddy").Funcs(FuncMap).Parse(caddyfile)
	if err != nil {
		log.Fatal("could not load template: ", zap.Error(err))
	}

	mainCaddy.caddyfile = caddyfile
	mainCaddy.caddytemplate = caddytemplate
	mainCaddy.player = player

	caddyLoader := func(serverType string) (caddy.Input, error) {
		buf, err := mainCaddy.Play()
		if err != nil {
			return nil, err
		}

		return caddy.CaddyfileInput{
			Contents:       buf.Bytes(),
			ServerTypeName: serverType,
		}, nil
	}

	caddy.SetDefaultCaddyfileLoader("http", caddy.LoaderFunc(caddyLoader))
}

// Get returns the currently enabled caddy builder
func Get() *Caddy {
	return mainCaddy
}

// Start caddy
func Start() error {
	// load caddyfile
	caddyfile, err := caddy.LoadCaddyfile("http")
	if err != nil {
		return err
	}

	LastKnownCaddyFile = string(caddyfile.Body())

	// start caddy server
	instance, err := caddy.Start(caddyfile)
	if err != nil {
		return err
	}

	mainCaddy.instance = instance
	return nil
}

func Stop() error {
	instance := GetInstance()
	instance.ShutdownCallbacks()
	instance.Stop()

	return nil
}

func StartWithFastRestart() (chan bool, error) {
	c := make(chan bool, 1)
	e := Start()
	go func() {
		defer close(c)
		<-time.After(2 * time.Second)

		log.Logger(context.Background()).Debug("Restarting Proxy Now (fast restart)")

		restart()
		restartRequired = false
	}()
	return c, e
}

func restart() error {

	if mainCaddy.instance == nil {
		return fmt.Errorf("instance not started")
	}

	// load caddyfile
	caddyfile, err := caddy.LoadCaddyfile("http")
	if err != nil {
		return err
	}

	LastKnownCaddyFile = string(caddyfile.Body())

	if common.LogLevel == zap.DebugLevel {
		fmt.Println(LastKnownCaddyFile)
	} else {
		log.Logger(gatewayCtx).Info("Restarting proxy", zap.ByteString("caddyfile", caddyfile.Body()))
	}

	// restart caddy server
	var instance *caddy.Instance
	if runtime.GOOS == "windows" {
		log.Logger(gatewayCtx).Info("Stopping Caddy Instance")
		if e := mainCaddy.instance.Stop(); e != nil {
			return e
		}
		mainCaddy.instance.ShutdownCallbacks()
		log.Logger(gatewayCtx).Info("Starting new Caddy Instance")
		instance, err = caddy.Start(caddyfile)
	} else {
		instance, err = mainCaddy.instance.Restart(caddyfile)
	}
	if err != nil {
		return err
	}

	log.Logger(gatewayCtx).Info("Restart done")

	mainCaddy.instance = instance

	broker.Publish(common.TopicProxyRestarted, &broker.Message{Body: []byte("")})

	return nil
}

func (c *Caddy) Play() (*bytes.Buffer, error) {
	return c.player()
}

func GetInstance() *caddy.Instance {
	return mainCaddy.instance
}

func (c *Caddy) GetTemplate() *template.Template {
	return c.caddytemplate
}

func ServiceReady(name string) bool {
	services, _ := registry.GetService(name)
	for _, service := range services {
		if len(service.Nodes) > 0 {
			return true
		}
	}
	return false
}

func internalURLFromServices(name string, uri ...string) string {
	var res []string

	services, _ := registry.GetService(name)

	for _, service := range services {
		for _, node := range service.Nodes {
			res = append(res, fmt.Sprintf("%s%s", node.Address, strings.Join(uri, "")))
		}
	}

	if len(res) == 0 {
		go func() {
			hooks.RestartChan <- true
		}()
		return "PENDING"
	}

	return strings.Join(res, " ")
}
