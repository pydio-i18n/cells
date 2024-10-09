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

package resources

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/storage/sql/resources"
	"github.com/pydio/cells/v4/common/telemetry/log"
)

type PoliciesCleanerOptions struct {
	SubscribeRoles bool
	SubscribeUsers bool
}

type PoliciesCleaner struct {
	Options PoliciesCleanerOptions
}

// Handle cleans resources in the current DAO based on the delete events
func (c *PoliciesCleaner) Handle(ctx context.Context, dao resources.DAO, msg *idm.ChangeEvent) error {

	if msg.Type != idm.ChangeEventType_DELETE {
		return nil
	}

	var subject string
	if c.Options.SubscribeRoles && msg.Role != nil {
		subject = fmt.Sprintf("role:%s", msg.Role.Uuid)
	}
	if c.Options.SubscribeUsers && msg.User != nil {
		if msg.User.IsGroup {
			subject = fmt.Sprintf("role:%s", msg.User.Uuid)
		} else {
			subject = fmt.Sprintf("user:%s", msg.User.Login)
		}
	}

	if len(subject) > 0 {
		log.Logger(ctx).Info("Deleting policies on event", zap.Any("event", msg), zap.String("subject", subject))
		return dao.DeletePoliciesBySubject(ctx, subject)
	}
	return nil

}
