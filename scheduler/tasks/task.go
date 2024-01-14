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

package tasks

import (
	"context"
	"google.golang.org/protobuf/encoding/protojson"
	"sync/atomic"
	"time"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/proto/jobs"
	"github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/utils/permissions"
	"github.com/pydio/cells/v4/common/utils/uuid"
	"github.com/pydio/cells/v4/scheduler/actions"
)

type Task struct {
	*jobs.Job
	common.RuntimeHolder
	runID string

	context  context.Context
	cancel   context.CancelFunc
	finished chan bool

	rci atomic.Int32

	chi int
	di  int

	event interface{}
	task  *jobs.Task

	lastStatus                    jobs.TaskStatus
	lastStatusMsg                 string
	lastHasProgress, lastCanPause bool
	lastProgress                  float32

	err error
}

// NewTaskFromEvent creates a task based on incoming job and event
func NewTaskFromEvent(runtime, ctx context.Context, job *jobs.Job, event interface{}) *Task {
	ctxUserName, _ := permissions.FindUserNameInContext(ctx)
	taskID := uuid.New()
	if trigger, ok := event.(*jobs.JobTriggerEvent); ok && trigger.RunTaskId != "" {
		taskID = trigger.RunTaskId
	}
	operationID := job.ID + "-" + taskID[0:8]
	c := servicecontext.WithOperationID(ctx, operationID)

	// Inject evaluated job parameters if it's not already here
	if len(job.Parameters) > 0 && c.Value(ContextJobParametersKey{}) == nil {
		params := make(map[string]string, len(job.Parameters))
		for _, p := range job.Parameters {
			params[p.Name] = jobs.EvaluateFieldStr(ctx, &jobs.ActionMessage{}, p.Value)
		}
		// Replace job parameters with values passed through TriggerEvent
		if jte, ok := event.(*jobs.JobTriggerEvent); ok && len(jte.RunParameters) > 0 {
			for k, v := range jte.RunParameters {
				if _, o := params[k]; o {
					params[k] = jobs.EvaluateFieldStr(ctx, &jobs.ActionMessage{}, v)
				}
			}
		}
		c = context.WithValue(c, ContextJobParametersKey{}, params)
	}

	t := &Task{
		context:  c,
		Job:      job,
		runID:    taskID,
		finished: make(chan bool, 1),
		event:    event,
		task: &jobs.Task{
			ID:            taskID,
			JobID:         job.ID,
			Status:        jobs.TaskStatus_Queued,
			StatusMessage: "Pending",
			ActionsLogs:   []*jobs.ActionLog{},
			TriggerOwner:  ctxUserName,
			CanStop:       true,
		},
	}
	t.SetRuntimeContext(runtime)
	return t
}

// Queue send this new task to the dispatcher queue.
// If a second queue is passed, it may differ from main input queue, so it is used for children queuing
func (t *Task) Queue(queue ...chan RunnerFunc) {
	var ct context.Context
	var can context.CancelFunc
	if d, o := itemTimeout(t.context, t.Job.Timeout); o {
		ct, can = context.WithTimeout(t.context, d)
	} else {
		ct, can = context.WithCancel(t.context)
	}
	t.context = ct
	t.cancel = can
	jobId := t.Job.ID
	taskId := t.runID

	ch := PubSub.Sub(PubSubTopicControl)
	go func() {
		defer func() {
			UnSubWithFlush(ch, PubSubTopicControl)
		}()
		for {
			select {
			case <-t.finished:
				return
			case <-t.RuntimeContext.Done():
				t.cancel()
			case val := <-ch:
				cmd, ok := val.(*jobs.CtrlCommand)
				if !ok {
					continue
				}
				if cmd.TaskId != "" && cmd.TaskId != taskId {
					continue
				}
				if cmd.JobId != "" && cmd.JobId != jobId {
					continue
				}
				if cmd.Cmd != jobs.Command_Stop {
					continue
				}
				t.cancel()
			}
		}
	}()
	r := RootRunnable(t.context, t)
	var secondaryQueue = queue[0]
	if len(queue) > 1 {
		secondaryQueue = queue[1]
	}
	if t.Job.MergeAction != nil {
		r.SetupCollector(t.context, t.Job.MergeAction, secondaryQueue)
	}
	logStartMessageFromEvent(r.Context, t.event)
	msg := createMessageFromEvent(t.event)
	queue[0] <- func(queue chan RunnerFunc) {
		r.Dispatch(msg, t.Actions, secondaryQueue)
	}
}

// CleanUp is triggered after a task has no more subroutines running.
func (t *Task) CleanUp() {
	t.SetEndTime(time.Now())
	if t.err != nil {
		t.SetStatus(jobs.TaskStatus_Error, t.err.Error())
	} else {
		t.SetStatus(jobs.TaskStatus_Finished, "Complete")
	}
	t.Save()
	close(t.finished)
}

// Add increments task internal retain counter
func (t *Task) Add(delta int) {
	rc := t.rci.Load()
	if rc == 0 {
		if t.task.StartTime == 0 {
			t.task.StartTime = int32(time.Now().Unix())
		}
		t.SetStatus(jobs.TaskStatus_Running, "Starting...")
		t.Save()
	}
	t.rci.Add(int32(delta))
}

// Done decrements task internal retain counter - When reaching 0, it triggers the CleanUp operation
func (t *Task) Done(delta int) {
	newVal := t.rci.Add(-int32(delta))
	if newVal == 0 {
		t.CleanUp()
	}
}

// Save publish task to PubSub topic
func (t *Task) Save() {
	if t.lastStatus == jobs.TaskStatus_Unknown || t.taskChanged() {
		cl := t.Clone()
		t.lastStatus = cl.Status
		t.lastStatusMsg = cl.StatusMessage
		t.lastHasProgress = cl.HasProgress
		t.lastCanPause = cl.CanPause
		t.lastProgress = cl.Progress
		PubSub.Pub(cl, PubSubTopicTaskStatuses)
	}
}

// Clone creates a protobuf clone of this task
func (t *Task) Clone() *jobs.Task {
	bb, _ := protojson.Marshal(t.task)
	cl := &jobs.Task{}
	_ = protojson.Unmarshal(bb, cl)
	return cl
	//return proto.Clone(t.task).(*jobs.Task)
}

// GetRunUUID returns the task internal run UUID
func (t *Task) GetRunUUID() string {
	return t.runID
}

// SetStatus updates task internal status
func (t *Task) SetStatus(status jobs.TaskStatus, message ...string) {
	if len(message) > 0 {
		t.task.StatusMessage = message[0]
	}
	t.task.Status = status
}

// SetProgress updates task internal progress
func (t *Task) SetProgress(progress float32) {
	t.task.Progress = progress
}

// SetStartTime updates start time
func (t *Task) SetStartTime(ti time.Time) {
	if t.task.StartTime == 0 {
		t.task.StartTime = int32(ti.Unix())
	}
}

// SetEndTime updates end time
func (t *Task) SetEndTime(ti time.Time) {
	t.task.EndTime = int32(ti.Unix())
}

// SetControllable flags task as being able to be stopped or paused
func (t *Task) SetControllable(canPause bool) {
	t.task.CanPause = canPause
}

// SetHasProgress flags task as providing progress information
func (t *Task) SetHasProgress() {
	t.task.HasProgress = true
}

// SetError set task in error globally
func (t *Task) SetError(e error, appendLog bool) {
	t.err = e
}

// GetRunnableChannels prepares a set of data channels for action actual Run method.
func (t *Task) GetRunnableChannels(controllable bool) (*actions.RunnableChannels, chan bool) {
	status, statusMsg, progress, done := t.createStatusesChannels()
	c := &actions.RunnableChannels{
		Status:    status,
		StatusMsg: statusMsg,
		Progress:  progress,
	}
	if controllable {
		c.Pause, c.Resume = t.createControlChannels(done)
	}
	return c, done
}

// createStatusesChannels provides a set of channel used by the runnable to send
// updates about its status to the outside world
func (t *Task) createStatusesChannels() (chan jobs.TaskStatus, chan string, chan float32, chan bool) {

	status := make(chan jobs.TaskStatus)
	statusMsg := make(chan string)
	progress := make(chan float32)
	done := make(chan bool, 1)

	go func() {
		defer func() {
			close(statusMsg)
			close(status)
			close(progress)
		}()
		for {
			select {
			case s := <-status:
				t.task.Status = s
				t.Save()
			case s := <-statusMsg:
				t.task.StatusMessage = s
				t.Save()
			case p := <-progress:
				diff := p - t.task.Progress
				save := false
				if diff > 0.01 || p == 1 {
					t.task.Progress = p
					save = true
				}
				if save {
					t.Save()
				}
			case <-done:
				return
			}

		}
	}()

	return status, statusMsg, progress, done

}

// createControlChannels provides a set of channel used to send some specific control instructions
// to the runnable
func (t *Task) createControlChannels(done chan bool) (pause chan interface{}, resume chan interface{}) {

	pause, resume = make(chan interface{}), make(chan interface{})
	jobId := t.Job.ID
	taskId := t.task.ID

	ch := PubSub.Sub(PubSubTopicControl)
	go func() {
		defer func() {
			close(pause)
			close(resume)
			UnSubWithFlush(ch, PubSubTopicControl)
		}()
		for {
			select {
			case val := <-ch:
				if cmd, ok := val.(*jobs.CtrlCommand); ok {
					if cmd.TaskId != "" && cmd.TaskId != taskId {
						continue
					}
					if cmd.JobId != "" && cmd.JobId != jobId {
						continue
					}
					switch cmd.Cmd {
					case jobs.Command_Pause:
						pause <- cmd
					case jobs.Command_Resume:
						resume <- cmd
					}
				}
			case <-done:
				return
			}
		}
	}()

	return
}

func (t *Task) taskChanged() bool {
	if t.lastStatus != t.task.Status || t.lastStatusMsg != t.task.StatusMessage {
		return true
	}
	if t.lastHasProgress != t.task.HasProgress || t.lastCanPause != t.task.CanPause || t.lastProgress != t.task.Progress {
		return true
	}
	return false
}
