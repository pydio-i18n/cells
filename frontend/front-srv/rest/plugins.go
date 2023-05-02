/*
 * Copyright (c) 2021. Abstrium SAS <team (at) pydio.com>
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

// Package rest is a service for serving specific requests directly to frontend
package rest

import (
	"context"
	"encoding/gob"
	"os"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/frontend"
	"github.com/pydio/cells/v4/common/service/frontend/sessions"
	"github.com/pydio/cells/v4/frontend/front-srv"
	"github.com/pydio/cells/v4/frontend/front-srv/rest/modifiers"
)

var BasePluginsBox = frontend.PluginBox{
	Box: front_srv.FrontendAssets,
	Exposes: []string{
		"access.directory",
		"access.gateway",
		"access.homepage",
		"access.settings",
		"action.avatar",
		"action.compression",
		"action.migration",
		"action.share",
		"action.user",
		"auth.pydio",
		"authfront.session_login",
		"core.activitystreams",
		"core.auth",
		"core.authfront",
		"core.conf",
		"core.mailer",
		"core.pydio",
		"core.uploader",
		"editor.browser",
		"editor.ckeditor",
		"editor.codemirror",
		"editor.diaporama",
		"editor.exif",
		"editor.infopanel",
		"editor.libreoffice",
		"editor.openlayer",
		"editor.pdfjs",
		"editor.soundmanager",
		"editor.text",
		"editor.video",
		"gui.ajax",
		"gui.mobile",
		"meta.comments",
		"meta.exif",
		"meta.simple_lock",
		"meta.user",
		"meta.versions",
		"uploader.html",
		"uploader.http",
		"uploader.uppy",
	},
}

func init() {

	if os.Getenv("CELLS_ENABLE_LIVEKIT") != "" {
		BasePluginsBox.Exposes = append(BasePluginsBox.Exposes, "action.livekit")
	}
	config.RegisterVaultKey("frontend/plugin/action.livekit/LK_API_SECRET")

	runtime.Register("main", func(ctx context.Context) {
		gob.Register(map[string]string{})

		//frontend.RegisterRegModifier(modifiers.MetaUserRegModifier)
		frontend.RegisterPluginModifier(modifiers.MetaUserPluginModifier)
		frontend.RegisterPluginModifier(modifiers.MobileRegModifier)

		frontend.WrapAuthMiddleware(modifiers.LogoutAuth)
		frontend.WrapAuthMiddleware(modifiers.RefreshAuth)

		frontend.WrapAuthMiddleware(modifiers.LoginPasswordAuth)
		frontend.WrapAuthMiddleware(modifiers.LoginExternalAuth)
		frontend.WrapAuthMiddleware(modifiers.AuthorizationCodeAuth)

		frontend.WrapAuthMiddleware(modifiers.LoginSuccessWrapper)
		frontend.WrapAuthMiddleware(modifiers.LoginFailedWrapper)

		service.NewService(
			service.Name(common.ServiceRestNamespace_+common.ServiceFrontend),
			service.Context(ctx),
			service.Tag(common.ServiceTagFrontend),
			service.Description("REST service for serving specific requests directly to frontend"),
			service.PluginBoxes(BasePluginsBox),
			service.WithStorage(sessions.NewDAO,
				service.WithStorageDefaultDriver(func() (string, string) {
					return "securecookie", ""
				}),
				service.WithStorageSupport("securecookie", "mysql"),
				service.WithStoragePrefix("idm_frontend_"),
			),
			service.WithWebSession("POST:/frontend/binaries"),
			service.WithWeb(func(c context.Context) service.WebHandler {
				dao := servicecontext.GetDAO(c)
				sessionDAO, ok := dao.(sessions.DAO)
				if !ok {
					panic("Cannot get SessionDAO")
				}
				// Depending on implementation, this will start a continuous background cleanup
				sessionDAO.DeleteExpired(c, log.Logger(c))
				return NewFrontendHandler(c, sessionDAO)
			}),
		)
	})

	if os.Getenv("CELLS_ENABLE_FORMS_DEVEL") == "1" {
		config.RegisterExposedConfigs("pydio.rest.forms-devel", formDevelConfigs)
	}

}
