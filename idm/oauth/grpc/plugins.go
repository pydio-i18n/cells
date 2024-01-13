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

// Package grpc spins an OpenID Connect Server using the coreos/dex implementation
package grpc

import (
	"context"
	"github.com/ory/fosite"
	"github.com/pydio/cells/v4/common/server"
	"log"

	log2 "github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/utils/configx"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/auth"
	"github.com/pydio/cells/v4/common/config"
	auth2 "github.com/pydio/cells/v4/common/proto/auth"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service"
	"github.com/pydio/cells/v4/idm/oauth"
)

var (
	Name = common.ServiceGrpcNamespace_ + common.ServiceOAuth
)

func init() {
	runtime.Register("main", func(ctx context.Context) {

		var srv grpc.ServiceRegistrar
		if !server.Get(&srv) {
			panic("no grpc server available")
		}

		h := &Handler{name: Name}
		auth2.RegisterAuthTokenVerifierServer(srv, h)
		auth2.RegisterLoginProviderServer(srv, h)
		auth2.RegisterConsentProviderServer(srv, h)
		auth2.RegisterLogoutProviderServer(srv, h)
		auth2.RegisterAuthCodeProviderServer(srv, h)
		auth2.RegisterAuthCodeExchangerServer(srv, h)
		auth2.RegisterAuthTokenRefresherServer(srv, h)
		auth2.RegisterAuthTokenRevokerServer(srv, h)
		auth2.RegisterAuthTokenPrunerServer(srv, h)
		auth2.RegisterPasswordCredentialsTokenServer(srv, h)

		watcher, _ := config.Watch(configx.WithPath("services", common.ServiceWebNamespace_+common.ServiceOAuth))
		go func() {
			for {
				values, er := watcher.Next()
				if er != nil {
					break
				}
				log2.Logger(ctx).Info("Reloading configurations for OAuth services")
				auth.InitConfiguration(values.(configx.Values))
			}
		}()

		// Registry initialization
		// Blocking on purpose, as it should block login
		//er := auth.InitRegistry(ctx, Name)
		//if er == nil {
		//	log2.Logger(ctx).Info("Finished auth.InitRegistry")
		//} else {
		//	log2.Logger(ctx).Info("Error while applying auth.InitRegistry", zap.Error(er))
		//}

		dao, err := oauth.NewDAO(ctx)
		if err != nil {
			panic(err)
		}

		opts := configx.New()
		dao.Init(ctx, opts)

		pat := NewPATHandler(dao.(oauth.DAO), &fosite.Config{})
		auth2.RegisterPersonalAccessTokenServiceServer(srv, pat)
		auth2.RegisterAuthTokenVerifierServer(srv, pat)
		auth2.RegisterAuthTokenPrunerServer(srv, pat)

		service.NewService(
			service.Name(Name),
			service.Context(ctx),
			service.Tag(common.ServiceTagAuth),
			service.Description("OAuth Provider"),
			//service.Migrations([]*service.Migration{
			//	{
			//		TargetVersion: service.FirstRun(),
			//		Up:            oauth.InsertPruningJob,
			//	},
			//}),
		)

		service.NewService(
			service.Name(common.ServiceGrpcNamespace_+common.ServiceToken),
			service.Context(ctx),
			service.Tag(common.ServiceTagIdm),
			service.Description("Personal Access Token Provider"),
			//service.WithTODOStorage(oauth.NewDAO, commonsql.NewDAO,
			//	service.WithStoragePrefix("idm_oauth_"),
			//	service.WithStorageSupport(mysql.Driver, sqlite.Driver),
			//),
			//service.WithGRPC(func(ctx context.Context, server grpc.ServiceRegistrar) error {
			//
			//	return nil
			//}),
		)

		auth.OnConfigurationInit(func(scanner configx.Scanner) {
			var m []struct {
				ID   string
				Name string
				Type string
			}

			if err := scanner.Scan(&m); err != nil {
				log.Fatal("Cannot correctly scan Connectors from config. JSON format maybe wrong, please check configuration. Error was ", err)
			}

			for _, mm := range m {
				if mm.Type == "pydio" {
					// Registering the first connector
					auth.RegisterConnector(mm.ID, mm.Name, mm.Type, nil)
				}
			}
		})

		// load configuration
		auth.InitConfiguration(config.Get("services", common.ServiceWebNamespace_+common.ServiceOAuth))

		// Register the services as GRPC Auth Providers
		auth.RegisterGRPCProvider(auth.ProviderTypeGrpc, common.ServiceGrpcNamespace_+common.ServiceOAuth)
		auth.RegisterGRPCProvider(auth.ProviderTypePAT, common.ServiceGrpcNamespace_+common.ServiceToken)
	})
}
