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

package timer

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/pydio/cells/v5/common"
	"github.com/pydio/cells/v5/common/broker"
	"github.com/pydio/cells/v5/common/client/grpc"
	"github.com/pydio/cells/v5/common/proto/jobs"
	"github.com/pydio/cells/v5/common/runtime"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/common/utils/schedule"
)

// EventProducer gathers all Tickers in a pool and provides a single entry point
// to communicate with them, typically to stop them.
type EventProducer struct {
	Context context.Context

	waiters   map[string]*schedule.Ticker
	waitersMu sync.Mutex

	eventChan chan *jobs.JobTriggerEvent
	stopChan  chan bool
	testChan  chan *jobs.JobTriggerEvent
}

// NewEventProducer creates a pool of ScheduleWaiters that will send events based on pre-defined scheduling.
func NewEventProducer(rootCtx context.Context) *EventProducer {
	e := &EventProducer{
		waiters:   make(map[string]*schedule.Ticker),
		waitersMu: sync.Mutex{},
		stopChan:  make(chan bool, 1),
		eventChan: make(chan *jobs.JobTriggerEvent),
	}

	rootCtx = runtime.WithServiceName(rootCtx, common.ServiceGrpcNamespace_+common.ServiceTimer)
	e.Context = context.WithValue(rootCtx, common.PydioContextUserKey, common.PydioSystemUsername)

	go func() {
		defer close(e.stopChan)
		defer close(e.eventChan)

		for {
			select {
			case event := <-e.eventChan:
				log.Logger(e.Context).Debug("Sending Timer Event", zap.Any("event", event))
				if e.testChan != nil {
					e.testChan <- event
				} else {
					broker.MustPublish(e.Context, common.TopicTimerEvent, event)
				}
			case <-rootCtx.Done():
				e.StopAll()
			case <-e.stopChan:
				return
			}
		}
	}()

	return e
}

// Start loads all TimersOnly Jobs from the job repository and registers them in this EventProducer pool.
func (e *EventProducer) Start() error {

	// Load all schedules
	cli := jobs.NewJobServiceClient(grpc.ResolveConn(e.Context, common.ServiceJobsGRPC))
	streamer, err := cli.ListJobs(e.Context, &jobs.ListJobsRequest{TimersOnly: true})
	if err != nil {
		return err
	}

	// Iterate through the registered jobs
	for {
		resp, err := streamer.Recv()
		if err != nil {
			break
		}
		if resp == nil {
			continue
		}
		log.Logger(e.Context).Info("Registering Job", zap.String("job", resp.Job.ID))
		e.StartOrUpdateJob(resp.Job)
	}
	return nil
}

// StopAll ranges all waiters from the EventProducer, calls Stop() and remove them from the Waiter pool.
func (e *EventProducer) StopAll() {
	e.waitersMu.Lock()
	for jId, w := range e.waiters {
		w.Stop()
		delete(e.waiters, jId)
	}
	e.waitersMu.Unlock()
	e.stopChan <- true
}

// StopWaiter stops a waiter given its ID and remove it from the Waiter pool.
// If no waiter with this ID is registered, it returns silently.
func (e *EventProducer) StopWaiter(jobId string) {
	e.waitersMu.Lock()
	defer e.waitersMu.Unlock()
	if w, ok := e.waiters[jobId]; ok {
		w.Stop()
		delete(e.waiters, jobId)
	}
}

// StartOrUpdateJob creates a ScheduleWaiter and registers it in the EventProducer pool.
// If a waiter already exists for the same job ID it will be stopped and replaced with a new one.
func (e *EventProducer) StartOrUpdateJob(job *jobs.Job) {

	// Stop if already running
	jobId := job.ID
	e.StopWaiter(jobId)

	//schedule := job.Schedule
	if s, err := schedule.NewTickerScheduleFromISO(job.Schedule.Iso8601Schedule); err == nil {
		w := schedule.NewTicker(s, func() error {
			e.eventChan <- &jobs.JobTriggerEvent{
				JobID:    jobId,
				Schedule: job.Schedule,
			}
			return nil
		})
		w.Start()
		e.waitersMu.Lock()
		e.waiters[jobId] = w
		e.waitersMu.Unlock()
	} else {
		log.Logger(e.Context).Error("Cannot register job", zap.Error(err))
	}
}

// Handle passes JobChangeEvents to the registered event producer.
func (e *EventProducer) Handle(ctx context.Context, msg *jobs.JobChangeEvent) error {

	log.Logger(ctx).Debug("JobsEvent Subscriber", zap.Any("event", msg))

	if msg.JobRemoved != "" {
		e.StopWaiter(msg.JobRemoved)
	}
	if msg.JobUpdated != nil && msg.JobUpdated.Schedule != nil {
		e.StartOrUpdateJob(msg.JobUpdated)
	}

	return nil
}
