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

package jobs

import (
	"context"
	"time"

	"github.com/pydio/cells/v5/common/proto/activity"
	"github.com/pydio/cells/v5/common/proto/idm"
	"github.com/pydio/cells/v5/common/proto/jobs"
	"github.com/pydio/cells/v5/common/proto/tree"
	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/common/service"
)

var Drivers = service.StorageDrivers{}

// DAO provides method interface to access the store for scheduler job and task definitions.
type DAO interface {
	PutJob(job *jobs.Job) error
	GetJob(jobId string, withTasks jobs.TaskStatus) (*jobs.Job, error)
	DeleteJob(jobId string) error
	ListJobs(owner string, eventsOnly bool, timersOnly bool, withTasks jobs.TaskStatus, jobIDs []string, taskCursor ...int32) (chan *jobs.Job, error)

	PutTask(task *jobs.Task) error
	PutTasks(task map[string]map[string]*jobs.Task) error
	ListTasks(jobId string, taskStatus jobs.TaskStatus, cursor ...int32) (chan *jobs.Task, chan bool, error)
	DeleteTasks(jobId string, taskId []string) error

	FindOrphans() ([]*jobs.Task, error)
	BuildOrphanLogsQuery(time.Duration, []string) string
}

// StripTaskData removes unnecessary data from the task log
// like fully loaded users, nodes, activities, etc.
func StripTaskData(task *jobs.Task) {
	for _, l := range task.ActionsLogs {
		if l.InputMessage != nil {
			stripTaskMessage(l.InputMessage)
		}
		if l.OutputMessage != nil {
			stripTaskMessage(l.OutputMessage)
		}
	}
}

// stripTaskMessage removes unnecessary data from the ActionMessage
func stripTaskMessage(message *jobs.ActionMessage) {
	for i, n := range message.Nodes {
		message.Nodes[i] = &tree.Node{Uuid: n.Uuid, Path: n.Path}
	}
	for i, u := range message.Users {
		message.Users[i] = &idm.User{Uuid: u.Uuid, Login: u.Login, GroupPath: u.GroupPath, GroupLabel: u.GroupLabel, IsGroup: u.IsGroup}
	}
	for i, a := range message.Activities {
		message.Activities[i] = &activity.Object{Id: a.Id}
	}
}

func Migrate(ctx, fromCtx, toCtx context.Context, dryRun bool, status chan service.MigratorStatus) (map[string]int, error) {

	out := map[string]int{
		"Jobs":  0,
		"Tasks": 0,
	}
	var from, to DAO
	var e error
	if from, e = manager.Resolve[DAO](fromCtx); e != nil {
		return nil, e
	}
	if to, e = manager.Resolve[DAO](toCtx); e != nil {
		return nil, e
	}

	jj, er := from.ListJobs("", false, false, jobs.TaskStatus_Any, []string{})
	if er != nil {
		return nil, er
	}
	for j := range jj {
		tasks := j.Tasks
		out["Jobs"]++
		out["Tasks"] += len(tasks)
		if dryRun {
			break
		}
		if e := to.PutJob(j); e != nil {
			return out, e
		}
		for _, ta := range tasks {
			if er := to.PutTask(ta); er != nil {
				return out, er
			}
		}
	}

	return out, nil
}
