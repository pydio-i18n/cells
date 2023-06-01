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

package oauth

import (
	"context"
	"fmt"
	"time"

	"github.com/pydio/cells/v4/common/service/errors"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/forms"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/auth"
	"github.com/pydio/cells/v4/common/proto/docstore"
	"github.com/pydio/cells/v4/common/proto/jobs"
	"github.com/pydio/cells/v4/common/utils/i18n"
	"github.com/pydio/cells/v4/idm/oauth/lang"
	"github.com/pydio/cells/v4/scheduler/actions"
)

func init() {
	actions.GetActionsManager().Register(pruneTokensActionName, func() actions.ConcreteAction {
		return &PruneTokensAction{}
	})
}

// InsertPruningJob adds a job to scheduler
func InsertPruningJob(ctx context.Context) error {

	log.Logger(ctx).Info("Inserting pruning job for revoked token and reset password tokens")

	T := lang.Bundle().GetTranslationFunc(i18n.GetDefaultLanguage(config.Get()))

	cli := jobs.NewJobServiceClient(grpc.GetClientConnFromCtx(ctx, common.ServiceJobs))
	if resp, e := cli.GetJob(ctx, &jobs.GetJobRequest{JobID: pruneTokensActionName}); e == nil && resp.Job != nil {
		return nil // Already exists
	} else if e != nil && errors.FromError(e).Code != 404 {
		log.Logger(ctx).Info("Insert pruning job: jobs service not ready yet :"+e.Error(), zap.Error(errors.FromError(e)))
		return e // not ready yet, retry
	}
	_, e := cli.PutJob(ctx, &jobs.PutJobRequest{Job: &jobs.Job{
		ID:    pruneTokensActionName,
		Owner: common.PydioSystemUsername,
		Label: T("Auth.PruneJob.Title"),
		Schedule: &jobs.Schedule{
			Iso8601Schedule: "R/2012-06-04T19:25:16.828696-07:00/PT60M", // Every hour
		},
		AutoStart:      false,
		MaxConcurrency: 1,
		Actions: []*jobs.Action{{
			ID: pruneTokensActionName,
		}},
	}})

	return e
}

var (
	pruneTokensActionName = "actions.auth.prune.tokens"
)

type PruneTokensAction struct {
	common.RuntimeHolder
}

func (c *PruneTokensAction) GetDescription(lang ...string) actions.ActionDescription {
	return actions.ActionDescription{
		ID:              pruneTokensActionName,
		Label:           "Prune tokens",
		Icon:            "delete-sweep",
		Category:        actions.ActionCategoryIDM,
		Description:     "Delete expired and revoked token from internal registry",
		SummaryTemplate: "",
		HasForm:         false,
		IsInternal:      true,
	}
}

func (c *PruneTokensAction) GetParametersForm() *forms.Form {
	return nil
}

// GetName Unique identifier
func (c *PruneTokensAction) GetName() string {
	return pruneTokensActionName
}

// Init pass parameters
func (c *PruneTokensAction) Init(job *jobs.Job, action *jobs.Action) error {
	return nil
}

// Run the actual action code
func (c *PruneTokensAction) Run(ctx context.Context, channels *actions.RunnableChannels, input *jobs.ActionMessage) (*jobs.ActionMessage, error) {

	T := lang.Bundle().GetTranslationFunc(i18n.GetDefaultLanguage(config.Get()))

	output := input

	// Prune revoked tokens on OAuth service
	cli := auth.NewAuthTokenPrunerClient(grpc.GetClientConnFromCtx(ctx, common.ServiceOAuth))
	if pruneResp, e := cli.PruneTokens(ctx, &auth.PruneTokensRequest{}); e != nil {
		return input.WithError(e), e
	} else {
		log.TasksLogger(ctx).Info("OAuth Service: " + T("Auth.PruneJob.Revoked", struct{ Count int32 }{Count: pruneResp.GetCount()}))
		output.AppendOutput(&jobs.ActionOutput{Success: true})
	}

	// Prune revoked tokens on OAuth service
	cli2 := auth.NewAuthTokenPrunerClient(grpc.GetClientConnFromCtx(ctx, common.ServiceToken))
	if pruneResp, e := cli2.PruneTokens(ctx, &auth.PruneTokensRequest{}); e != nil {
		return input.WithError(e), e
	} else {
		log.TasksLogger(ctx).Info("Token Service: " + T("Auth.PruneJob.Revoked", struct{ Count int32 }{Count: pruneResp.GetCount()}))
		output.AppendOutput(&jobs.ActionOutput{Success: true})
	}

	// Prune reset password tokens
	docCli := docstore.NewDocStoreClient(grpc.GetClientConnFromCtx(ctx, common.ServiceDocStore))
	deleteResponse, er := docCli.DeleteDocuments(ctx, &docstore.DeleteDocumentsRequest{
		StoreID: "resetPasswordKeys",
		Query: &docstore.DocumentQuery{
			MetaQuery: fmt.Sprintf("expiration<%d", time.Now().Unix()),
		},
	})
	if er != nil {
		return output.WithError(er), er
	} else {
		log.TasksLogger(ctx).Info(T("Auth.PruneJob.ResetToken", deleteResponse))
		output.AppendOutput(&jobs.ActionOutput{Success: true})
	}

	return output, nil
}
