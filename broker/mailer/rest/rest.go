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

package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/pydio/cells/v4/common/client/grpc"
	"regexp"

	restful "github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/auth/claim"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/mailer"
	"github.com/pydio/cells/v4/common/service"
	"github.com/pydio/cells/v4/common/utils/i18n"
	"github.com/pydio/cells/v4/common/utils/permissions"
)

var (
	ErrBadFormat = errors.New("invalid format")

	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func ValidateFormat(email string) error {
	if !emailRegexp.MatchString(email) {
		return ErrBadFormat
	}
	return nil
}

// MailerHandler provides implementation of method interfaces
// to communicate with the configured MTA for this instance
type MailerHandler struct {
	RuntimeCtx context.Context
}

// SwaggerTags list the names of the service tags declared in the swagger json implemented by this service
func (mh *MailerHandler) SwaggerTags() []string {
	return []string{"MailerService"}
}

// Filter returns a function to filter the swagger path
func (mh *MailerHandler) Filter() func(string) string {
	return nil
}

// Send puts a mail in the queue to be send
func (mh *MailerHandler) Send(req *restful.Request, rsp *restful.Response) {

	var message mailer.Mail
	req.ReadEntity(&message)

	ctx := req.Request.Context()
	if len(message.To) > 100 {
		service.RestError403(req, rsp, fmt.Errorf("you are not allowed to send emails to more than 100 people at once"))
	}
	log.Logger(ctx).Debug("Sending Email", log.DangerouslyZapSmallSlice("to", message.To), zap.String("subject", message.Subject), zap.Any("templateData", message.TemplateData))

	langs := i18n.UserLanguagesFromRestRequest(req, config.Get())
	cli := mailer.NewMailerServiceClient(grpc.GetClientConnFromCtx(mh.RuntimeCtx, common.ServiceMailer))

	claims, ok := ctx.Value(claim.ContextKey).(claim.Claims)
	if !ok {
		service.RestError500(req, rsp, fmt.Errorf("sending email anonymously is forbidden"))
		return
	}
	// Safe defaults, but maybe reloaded below
	message.From = &mailer.User{
		Uuid:    claims.Subject, // user UUID
		Name:    claims.Name,
		Address: claims.Email,
	}

	// Reload user, as his displayName/email may have changed during session
	if u, er := permissions.SearchUniqueUser(ctx, "", claims.Subject); er != nil {
		service.RestError500(req, rsp, er)
		return
	} else if email, has := u.GetAttributes()["email"]; has {
		message.From.Address = email
		if display, has := u.GetAttributes()["displayName"]; has {
			message.From.Name = display
		} else {
			message.From.Name = u.GetLogin()
		}
	}

	var resolvedTos []*mailer.User
	for _, to := range message.To {
		if resolved, e := mh.ResolveUser(ctx, to); e == nil {
			if resolved.Language == "" && len(langs) > 0 {
				resolved.Language = langs[0]
			}
			resolvedTos = append(resolvedTos, resolved)
		} else {
			log.Logger(ctx).Error("ignoring sendmail for user as no email was found", zap.Any("user", to))
		}
	}
	if len(resolvedTos) == 0 {
		service.RestError500(req, rsp, fmt.Errorf("could not find any address to send to"))
		return
	}
	message.To = resolvedTos
	queue := true
	if message.TemplateId == "AdminTestMail" {
		queue = false
		message.TemplateData["AdminName"] = claims.Name
	}

	// Now call service to send email
	response, err := cli.SendMail(ctx, &mailer.SendMailRequest{Mail: &message, InQueue: queue})
	if err != nil {
		log.Logger(ctx).Error("could not send mail", zap.Error(err))
		service.RestError500(req, rsp, err)
		return
	}
	response.Success = true // make sure success is set
	rsp.WriteEntity(response)
}

func (mh *MailerHandler) ResolveUser(ctx context.Context, user *mailer.User) (*mailer.User, error) {
	if user.Address != "" {
		return user, nil
	}
	emailOrAddress := user.Uuid
	// Check if it's a user Login
	if u, e := permissions.SearchUniqueUser(ctx, emailOrAddress, ""); e == nil && u != nil {
		if email, has := u.GetAttributes()["email"]; has {
			output := &mailer.User{Uuid: u.GetUuid(), Address: email}
			if display, has := u.GetAttributes()["displayName"]; has {
				output.Name = display
			} else {
				output.Name = emailOrAddress
			}
			output.Language = i18n.UserLanguage(ctx, u, config.Get())
			return output, nil
		} else {
			return nil, fmt.Errorf("user %s has no email set", emailOrAddress)
		}
	} else if testErr := ValidateFormat(emailOrAddress); testErr == nil {
		return &mailer.User{
			Uuid:    emailOrAddress,
			Address: emailOrAddress,
			Name:    emailOrAddress,
		}, nil
	}
	return nil, fmt.Errorf("could not find any address for the passed uuid")
}
