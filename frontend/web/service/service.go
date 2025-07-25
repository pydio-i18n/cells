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

// Package web is serving the main entry points for the JS frontend
package service

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/lpar/gzipped"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/broker"
	"github.com/pydio/cells/v5/common/config"
	"github.com/pydio/cells/v5/common/config/routing"
	"github.com/pydio/cells/v5/common/middleware"
	"github.com/pydio/cells/v5/common/proto/front"
	"github.com/pydio/cells/v5/common/runtime"
	"github.com/pydio/cells/v5/common/service"
	"github.com/pydio/cells/v5/common/service/frontend"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/frontend/web"
)

const (
	Name         = common.ServiceWebNamespace_ + common.ServiceFrontStatics
	RobotsString = `User-agent: *
Disallow: /`
	ResetPasswordPath = "/user/reset-password/"
)

type recoveryLogger struct {
	logger log.ZapLogger
}

func (r *recoveryLogger) Println(vv ...interface{}) {
	var fields []zap.Field
	for _, v := range vv {
		switch t := v.(type) {
		case error:
			fields = append(fields, zap.Error(t))
		default:
			fields = append(fields, zap.Any("value", v))
		}
	}
	r.logger.Error("http error recovered", fields...)
}

func init() {

	routing.RegisterRoute(common.RouteFrontend, "Main Frontend", common.DefaultRouteFrontend)
	routing.RegisterRoute(common.RoutePublic, "Public links access", common.DefaultRoutePublic)

	runtime.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(common.ServiceGrpcNamespace_+common.ServiceFrontStatics),
			service.Context(ctx),
			service.Tag(common.ServiceTagFrontend),
			service.Description("Grpc service for internal requests about frontend manifest"),
			service.WithGRPC(func(ctx context.Context, server grpc.ServiceRegistrar) error {
				mH := &web.ManifestHandler{HandlerName: common.ServiceGrpcNamespace_ + common.ServiceFrontStatics}
				front.RegisterManifestServiceServer(server, mH)
				return nil
			}),
		)
		service.NewService(
			service.Name(Name),
			service.Context(ctx),
			service.Tag(common.ServiceTagFrontend),
			service.Description("WEB service for serving statics"),
			service.Migrations([]*service.Migration{
				{
					TargetVersion: service.FirstRun(),
					Up: func(ctx context.Context) error {
						data := map[string]interface{}{
							"secureHeaders": map[string]interface{}{
								"X-XSS-Protection": "1; mode=block",
							},
							"plugin": map[string]interface{}{
								"editor.libreoffice": map[string]interface{}{
									"LIBREOFFICE_HOST": "localhost",
									"LIBREOFFICE_PORT": "9980",
									"LIBREOFFICE_SSL":  true,
								},
							},
						}
						return config.Set(ctx, data, "frontend")
					},
				},
				{
					TargetVersion: service.ValidVersion("1.2.0"),
					Up:            DropLegacyStatics,
				},
			}),
			service.WithHTTP(func(ctx context.Context, mux routing.RouteRegistrar) error {

				timeoutWrap := func(handler http.Handler) http.Handler {
					return http.TimeoutHandler(handler, 15*time.Second, "There was a timeout while serving the frontend resources...")
				}
				recoveryWrap := handlers.RecoveryHandler(handlers.RecoveryLogger(&recoveryLogger{logger: log.Logger(ctx)}))

				httpFs := http.FS(frontend.GetPluginsFS())
				fs := gzipped.FileServer(httpFs)
				fs = timeoutWrap(fs)
				fs = recoveryWrap(fs)

				m := mux.Route(common.RouteFrontend)
				m.Handle("index.json", fs)
				m.Handle("plug/", fs, routing.WithStripPrefix())
				indexHandler := web.NewIndexHandler(ctx, ResetPasswordPath)
				indexHandler = middleware.HttpContextWrapper(ctx, indexHandler)
				indexHandler = timeoutWrap(indexHandler)
				indexHandler = recoveryWrap(indexHandler)

				m.Handle("robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Header().Set("Content-Type", "text/plain")
					_, _ = w.Write([]byte(RobotsString))
				}))
				m.Handle("/", indexHandler, routing.WithRewriteCatchAll())
				m.Handle(ResetPasswordPath, indexHandler)

				// /public endpoint : special handler for index, redirect to /plug/ for the rest
				ph := web.NewPublicHandler(ctx)
				handler := middleware.HttpContextWrapper(ctx, ph)
				handler = timeoutWrap(handler)
				handler = recoveryWrap(handler)

				pub := mux.Route(common.RoutePublic)
				pub.Handle("/", handler)
				pub.Handle("plug/", fs, routing.WithStripPrefix())

				// Adding subscriber
				_ = broker.SubscribeCancellable(ctx, common.TopicReloadAssets, func(_ context.Context, message broker.Message) error {
					log.Logger(ctx).Info("Reloading frontend plugins from file system")
					frontend.HotReload()
					httpFs = http.FS(frontend.GetPluginsFS())
					return nil
				}, broker.WithCounterName("frontend"))

				return nil
			}),
			service.WithHTTPStop(func(ctx context.Context, reg routing.RouteRegistrar) error {
				reg.DeregisterRoute(common.RouteFrontend)
				reg.DeregisterRoute(common.RoutePublic)
				return nil
			}),
		)
	})
}

// DropLegacyStatics removes files and references to old PHP data in configuration
func DropLegacyStatics(ctx context.Context) error {

	frontRoot := config.Get(ctx, "defaults", "frontRoot").Default(filepath.Join(runtime.ApplicationWorkingDir(), "static", "pydio")).String()
	if frontRoot != "" {
		if er := os.RemoveAll(frontRoot); er != nil {
			log.Logger(ctx).Error("Could not remove old PHP data from "+frontRoot+". You may safely delete this folder. Error was", zap.Error(er))
		} else {
			log.Logger(ctx).Info("Successfully removed old PHP data from " + frontRoot)
		}
	}

	log.Logger(ctx).Info("Clearing unused configurations")
	config.Del(ctx, "defaults", "frontRoot")
	config.Del(ctx, "defaults", "fpm")
	config.Del(ctx, "defaults", "fronts")
	config.Del(ctx, "services", "pydio.frontends")
	if config.Get(ctx, config.FrontendPluginPath(config.KeyFrontPluginCorePydio, config.KeyFrontApplicationTitle)...).String() == "" {
		_ = config.Set(ctx, "Pydio Cells", config.FrontendPluginPath(config.KeyFrontPluginCorePydio, config.KeyFrontApplicationTitle)...)
	}
	if e := config.Save(ctx, common.PydioSystemUsername, "Upgrade to 1.2.0"); e == nil {
		log.Logger(ctx).Info("[Upgrade] Cleaned unused configurations")
	}

	return nil
}
