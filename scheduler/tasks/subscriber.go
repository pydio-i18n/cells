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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cskr/pubsub"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	"github.com/pydio/cells/v4/common/proto/jobs"
	"github.com/pydio/cells/v4/common/proto/object"
	rpb "github.com/pydio/cells/v4/common/proto/registry"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/runtime"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/utils/permissions"
	"github.com/pydio/cells/v4/common/utils/queue"
	"github.com/pydio/cells/v4/common/utils/std"
)

const (
	PubSubTopicTaskStatuses = "tasks"
	PubSubTopicControl      = "control"
)

var (
	PubSub *pubsub.PubSub
)

// UnSubWithFlush wraps PubSub.Unsub with a select to make sure all messages are consumed before unsubscribing.
func UnSubWithFlush(ch chan interface{}, topics ...string) {
consume:
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				break consume
			}
			//fmt.Println("Unsub", topics, "there was still something to consume...")
		case <-time.After(3 * time.Second):
			//fmt.Println("Unsub", topics, "Break loop...")
			break consume
		}
	}
	PubSub.Unsub(ch, topics...)
}

type ContextJobParametersKey struct{}

// Subscriber handles incoming events, applies selectors if any
// and generates all ActionMessage to trigger actions
type Subscriber struct {
	rootCtx context.Context

	sync.RWMutex
	definitions map[string]*jobs.Job

	dispatcherLock sync.Mutex
	dispatchers    map[string]*Dispatcher
}

// NewSubscriber creates a multiplexer for tasks managements and messages
// by maintaining a map of dispatcher, one for each job definition.
func NewSubscriber(parentContext context.Context) *Subscriber {

	s := &Subscriber{
		definitions: make(map[string]*jobs.Job),
		//queue:       make(chan Runnable),
		dispatchers: make(map[string]*Dispatcher),
	}

	PubSub = pubsub.New(0)

	s.rootCtx = context.WithValue(parentContext, common.PydioContextUserKey, common.PydioSystemUsername)
	treeQu, er := queue.OpenQueue(s.rootCtx, runtime.PersistingQueueURL("serviceName", common.ServiceGrpcNamespace_+common.ServiceTasks, "name", common.TopicTreeChanges))
	if er != nil {
		log.Logger(s.rootCtx).Error("Cannot start treeQueue, using an in-memory instead", zap.Error(er))
		treeQu, _ = queue.OpenQueue(s.rootCtx, runtime.QueueURL("debounce", "2s", "idle", "20s", "max", "2000"))
	}
	metaQu, er := queue.OpenQueue(s.rootCtx, runtime.PersistingQueueURL("serviceName", common.ServiceGrpcNamespace_+common.ServiceTasks, "name", common.TopicMetaChanges))
	if er != nil {
		log.Logger(s.rootCtx).Error("Cannot start metaQueue, using an in-memory instead", zap.Error(er))
		metaQu, _ = queue.OpenQueue(s.rootCtx, runtime.QueueURL("debounce", "2s", "idle", "20s", "max", "2000"))
	}

	queueOpt := broker.Queue("tasks")
	counterOpt := broker.WithCounterName("tasks")

	_ = broker.SubscribeCancellable(parentContext, common.TopicTreeChanges, func(message broker.Message) error {
		md, bb := message.RawData()
		event := &tree.NodeChangeEvent{}
		if e := proto.Unmarshal(bb, event); e == nil {
			// Ignore events on Temporary nodes and internal nodes and optimistic
			if event.Optimistic {
				return nil
			}
			if event.Target != nil && (event.Target.Etag == common.NodeFlagEtagTemporary || event.Target.HasMetaKey(common.MetaNamespaceDatasourceInternal)) {
				return nil
			}
			if event.Type == tree.NodeChangeEvent_DELETE && event.Source.HasMetaKey(common.MetaNamespaceDatasourceInternal) {
				return nil
			}
			s.processNodeEvent(metadata.NewContext(s.rootCtx, md), event)
			return nil
		} else {
			return e
		}
	}, queueOpt, broker.WithLocalQueue(treeQu), counterOpt)

	_ = broker.SubscribeCancellable(parentContext, common.TopicTimerEvent, func(message broker.Message) error {
		target := &jobs.JobTriggerEvent{}
		if ctx, e := message.Unmarshal(target); e == nil {
			return s.timerEvent(ctx, target)
		}
		return nil
	}, queueOpt, counterOpt)

	_ = broker.SubscribeCancellable(parentContext, common.TopicJobConfigEvent, func(message broker.Message) error {
		js := &jobs.JobChangeEvent{}
		if ctx, e := message.Unmarshal(js); e == nil {
			return s.jobsChangeEvent(ctx, js)
		}
		return nil
	}, queueOpt, counterOpt)

	_ = broker.SubscribeCancellable(parentContext, common.TopicMetaChanges, func(message broker.Message) error {
		target := &tree.NodeChangeEvent{}
		md, bb := message.RawData()
		if e := proto.Unmarshal(bb, target); e == nil && (target.Type == tree.NodeChangeEvent_UPDATE_META || target.Type == tree.NodeChangeEvent_UPDATE_USER_META) {
			s.processNodeEvent(metadata.NewContext(s.rootCtx, md), target)
		}
		return nil
	}, queueOpt, broker.WithLocalQueue(metaQu), counterOpt)

	_ = broker.SubscribeCancellable(parentContext, common.TopicIdmEvent, func(message broker.Message) error {
		target := &idm.ChangeEvent{}
		if ctx, e := message.Unmarshal(target); e == nil {
			return s.idmEvent(ctx, target)
		}
		return nil
	}, queueOpt, counterOpt)

	//s.listenToQueue()
	s.taskChannelSubscription()

	return s
}

// Init subscriber with current list of jobs from Jobs service
func (s *Subscriber) Init(ctx context.Context) error {

	// Load Jobs Definitions
	jobClients := jobs.NewJobServiceClient(grpc.GetClientConnFromCtx(s.rootCtx, common.ServiceJobs))
	streamer, e := jobClients.ListJobs(ctx, &jobs.ListJobsRequest{})
	if e != nil {
		return e
	}
	s.Lock()
	s.dispatcherLock.Lock()
	defer s.Unlock()
	defer s.dispatcherLock.Unlock()
	for {
		resp, er := streamer.Recv()
		if er != nil {
			break
		}
		if resp == nil {
			continue
		}
		if resp.Job.Inactive {
			continue
		}
		s.definitions[resp.Job.ID] = resp.Job
		s.getDispatcherForJob(resp.Job, false)
	}

	return nil

}

// Stop closes internal EventsBatcher
func (s *Subscriber) Stop() {
	s.dispatcherLock.Lock()
	for _, d := range s.dispatchers {
		d.Stop()
	}
	s.dispatcherLock.Unlock()
}

func (s *Subscriber) enqueue(ctx context.Context, job *jobs.Job, event proto.Message) {
	dispatcher := s.getDispatcherForJob(job, true)
	if dispatcher.fifo != nil {
		_ = dispatcher.fifo.Push(ctx, event)
	} else {
		task := NewTaskFromEvent(s.rootCtx, ctx, job, event)
		task.Queue(dispatcher.Queue())
	}
}

// taskChannelSubscription uses PubSub library to receive update messages from tasks
func (s *Subscriber) taskChannelSubscription() {
	ch := PubSub.Sub(PubSubTopicTaskStatuses)
	cli := NewTaskReconnectingClient(s.rootCtx)
	cli.StartListening(ch)
}

// getDispatcherForJob creates a new dispatcher for a job
func (s *Subscriber) getDispatcherForJob(job *jobs.Job, lock bool) *Dispatcher {

	if lock {
		s.dispatcherLock.Lock()
		defer s.dispatcherLock.Unlock()
	}

	if d, exists := s.dispatchers[job.ID]; exists {
		return d
	}
	maxWorkers := DefaultMaximumWorkers
	if job.MaxConcurrency > 0 {
		maxWorkers = int(job.MaxConcurrency)
	}
	tags := map[string]string{
		"service": common.ServiceGrpcNamespace_ + common.ServiceTasks,
		"jobID":   job.ID,
	}
	dispatcher := NewDispatcher(s.rootCtx, maxWorkers, job, tags)
	s.dispatchers[job.ID] = dispatcher
	dispatcher.Run()
	return dispatcher
}

// Job Configuration was updated, react accordingly
func (s *Subscriber) jobsChangeEvent(ctx context.Context, msg *jobs.JobChangeEvent) error {
	s.Lock()
	s.dispatcherLock.Lock()
	// Update config
	if msg.JobRemoved != "" {
		delete(s.definitions, msg.JobRemoved)
		if dispatcher, ok := s.dispatchers[msg.JobRemoved]; ok {
			dispatcher.Stop()
			delete(s.dispatchers, msg.JobRemoved)
		}
	}
	if msg.JobUpdated != nil {
		s.definitions[msg.JobUpdated.ID] = msg.JobUpdated
		if dispatcher, ok := s.dispatchers[msg.JobUpdated.ID]; ok {
			dispatcher.Stop()
			delete(s.dispatchers, msg.JobUpdated.ID)
			if !msg.JobUpdated.Inactive {
				s.getDispatcherForJob(msg.JobUpdated, false)
			}
		}
	}
	s.dispatcherLock.Unlock()
	s.Unlock()
	// AutoStart if required
	if msg.JobUpdated != nil && !msg.JobUpdated.Inactive && msg.JobUpdated.AutoStart {
		if e := s.timerEvent(ctx, &jobs.JobTriggerEvent{JobID: msg.JobUpdated.ID, RunNow: true}); e != nil {
			log.Logger(s.rootCtx).Error("Cannot trigger job "+msg.JobUpdated.GetLabel()+" on AutoStart after update", zap.Error(e))
		} else {
			log.Logger(s.rootCtx).Info("AutoStarting Job " + msg.JobUpdated.GetLabel() + " after update")
		}
	}

	return nil
}

// prepareTaskContext creates adequate context for launching a task
func (s *Subscriber) prepareTaskContext(ctx context.Context, job *jobs.Job, addSystemUser bool, eventParameters ...map[string]string) context.Context {

	// Add System User if necessary
	if addSystemUser {
		if u, _ := permissions.FindUserNameInContext(ctx); u == "" {
			ctx = metadata.WithAdditionalMetadata(ctx, map[string]string{common.PydioContextUserKey: common.PydioSystemUsername})
			ctx = context.WithValue(ctx, common.PydioContextUserKey, common.PydioSystemUsername)
		}
	}

	md, ok := metadata.FromContextCopy(ctx)
	ctx = runtime.ForkContext(ctx, s.rootCtx)
	if ok {
		ctx = metadata.NewContext(ctx, md)
	}

	// Inject evaluated job parameters
	if len(job.Parameters) > 0 {
		params := make(map[string]string, len(job.Parameters))
		for _, p := range job.Parameters {
			params[p.Name] = jobs.EvaluateFieldStr(ctx, &jobs.ActionMessage{}, p.Value)
		}
		if len(eventParameters) > 0 {
			// Replace job parameters with values passed through TriggerEvent
			for k, v := range eventParameters[0] {
				if _, o := params[k]; o {
					params[k] = jobs.EvaluateFieldStr(ctx, &jobs.ActionMessage{}, v)
				}
			}
		}
		ctx = context.WithValue(ctx, ContextJobParametersKey{}, params)
	}

	return ctx
}

// timerEvent reacts to a trigger sent by the timer service
func (s *Subscriber) timerEvent(ctx context.Context, event *jobs.JobTriggerEvent) error {
	jobId := event.JobID
	// Load Job Data, build selectors
	s.Lock()
	defer s.Unlock()
	j, ok := s.definitions[jobId]
	if !ok {
		// Try to load definition directly for JobsService
		jobClients := jobs.NewJobServiceClient(grpc.GetClientConnFromCtx(s.rootCtx, common.ServiceJobs))
		resp, e := jobClients.GetJob(ctx, &jobs.GetJobRequest{JobID: jobId})
		if e != nil || resp.Job == nil {
			return nil
		}
		j = resp.Job
		// Shall we prepare dispatcher  ?
		// s.getDispatcherForJob(j, false)
	}
	if j.Inactive {
		return nil
	}
	if err := s.requiresUnsupportedCapacity(ctx, j); err != nil {
		return nil
	}
	if event.GetRunNow() && event.GetRunParameters() != nil {
		ctx = s.prepareTaskContext(ctx, j, true, event.GetRunParameters())
	} else {
		ctx = s.prepareTaskContext(ctx, j, true)
	}
	if event.GetRunNow() {
		log.Logger(ctx).Info("Run Job " + jobId + " on demand")
	} else {
		log.Logger(ctx).Info("Run Job " + jobId + " on timer event " + event.Schedule.String())
	}
	s.enqueue(ctx, j, event)

	return nil
}

// nodeEvent reacts to a trigger linked to a nodeChange event.
func (s *Subscriber) nodeEvent(ctx context.Context, event *tree.NodeChangeEvent) error {

	if event.Optimistic {
		return nil
	}

	// Always ignore events on Temporary nodes and internal nodes
	if event.Target != nil && (event.Target.Etag == common.NodeFlagEtagTemporary || event.Target.HasMetaKey(common.MetaNamespaceDatasourceInternal)) {
		return nil
	}
	if event.Type == tree.NodeChangeEvent_DELETE && event.Source.HasMetaKey(common.MetaNamespaceDatasourceInternal) {
		return nil
	}

	return nil
}

// processNodeEvent actually process batched events
func (s *Subscriber) processNodeEvent(ctx context.Context, event *tree.NodeChangeEvent) {

	s.Lock()
	defer s.Unlock()

	for jobId, jobData := range s.definitions {
		if jobData.Inactive {
			continue
		}
		sameJobUuid := s.contextJobSameUuid(ctx, jobId)
		tCtx := s.prepareTaskContext(ctx, jobData, false)
		var eventMatch string
		for _, eName := range jobData.EventNames {
			if eType, ok := jobs.ParseNodeChangeEventName(eName); ok {
				if event.Type == eType {
					if sameJobUuid {
						log.Logger(tCtx).Debug("Preventing loop for job " + jobData.Label + " on event " + eName)
						continue
					}
					eventMatch = eName
					break
				}
			}
		}
		if eventMatch == "" {
			continue
		}
		if err := s.requiresUnsupportedCapacity(ctx, jobData); err != nil {
			continue
		}
		if jobData.ContextMetaFilter != nil && !s.jobLevelContextFilterPass(tCtx, jobData.ContextMetaFilter) {
			continue
		}
		if jobData.NodeEventFilter != nil && !s.jobLevelFilterPass(tCtx, event, jobData.NodeEventFilter) {
			continue
		}
		if jobData.IdmFilter != nil && !s.jobLevelIdmFilterPass(tCtx, createMessageFromEvent(event), jobData.IdmFilter) {
			continue
		}
		if jobData.DataSourceFilter != nil && !s.jobLevelDataSourceFilterPass(ctx, event, jobData.DataSourceFilter) {
			continue
		}

		log.Logger(tCtx).Debug("Run Job " + jobId + " on event " + eventMatch)
		s.enqueue(tCtx, jobData, event)
	}

}

// idmEvent Reacts to a trigger linked to a nodeChange event.
func (s *Subscriber) idmEvent(ctx context.Context, event *idm.ChangeEvent) error {

	s.Lock()
	defer s.Unlock()

	for jobId, jobData := range s.definitions {
		if jobData.Inactive {
			continue
		}
		sameJob := s.contextJobSameUuid(ctx, jobId)
		tCtx := s.prepareTaskContext(ctx, jobData, true)
		if jobData.ContextMetaFilter != nil && !s.jobLevelContextFilterPass(tCtx, jobData.ContextMetaFilter) {
			continue
		}
		if jobData.IdmFilter != nil && !s.jobLevelIdmFilterPass(tCtx, createMessageFromEvent(event), jobData.IdmFilter) {
			continue
		}
		for _, eName := range jobData.EventNames {
			if jobs.MatchesIdmChangeEvent(eName, event) {
				if sameJob {
					log.Logger(tCtx).Debug("Prevent loop for job " + jobData.Label + " on event " + eName)
					continue
				}
				if err := s.requiresUnsupportedCapacity(ctx, jobData); err != nil {
					continue
				}
				log.Logger(tCtx).Debug("Run Job " + jobId + " on event " + eName)
				s.enqueue(tCtx, jobData, event)
			}
		}
	}
	return nil
}

// jobLevelFilterPass checks if a node must go through jobs at all (if there is a NodesSelector on the job level)
func (s *Subscriber) jobLevelFilterPass(ctx context.Context, event *tree.NodeChangeEvent, filter *jobs.NodesSelector) bool {
	var refNode *tree.Node
	if event.Target != nil {
		refNode = event.Target
	} else if event.Source != nil {
		refNode = event.Source
	}
	if refNode == nil {
		return true // Ignore
	}
	input := &jobs.ActionMessage{Nodes: []*tree.Node{refNode}}
	_, _, pass := filter.Filter(ctx, input)
	return pass
}

// jobLevelIdmFilterPass tests filter and return false if all input IDM slots are empty
func (s *Subscriber) jobLevelIdmFilterPass(ctx context.Context, input *jobs.ActionMessage, filter *jobs.IdmSelector) bool {
	_, _, pass := filter.Filter(ctx, input)
	return pass
}

// jobLevelContextFilterPass tests filter and return false if context is filtered out
func (s *Subscriber) jobLevelContextFilterPass(ctx context.Context, filter *jobs.ContextMetaFilter) bool {
	_, _, pass := filter.Filter(ctx, &jobs.ActionMessage{})
	return pass
}

// jobLevelDataSourceFilterPass tests filter and return false if datasource is filtered out
func (s *Subscriber) jobLevelDataSourceFilterPass(ctx context.Context, event *tree.NodeChangeEvent, filter *jobs.DataSourceSelector) bool {
	var refNode *tree.Node
	if event.Target != nil {
		refNode = event.Target
	} else if event.Source != nil {
		refNode = event.Source
	}
	if refNode == nil {
		return true // Ignore
	}
	if dsName := refNode.GetStringMeta(common.MetaNamespaceDatasourceName); dsName != "" {
		var ds *object.DataSource
		e := std.Retry(ctx, func() error {
			var er error
			ds, er = config.GetSourceInfoByName(dsName)
			return er
		}, 3, 20)
		if e != nil {
			log.Logger(ctx).Error("jobLevelDataSourceFilter : cannot load source by name " + dsName + " - Job will not run!")
			return false
		}
		_, _, pass := filter.Filter(ctx, &jobs.ActionMessage{DataSources: []*object.DataSource{ds}})
		log.Logger(ctx).Debug("Filtering on node datasource (from node meta)", zap.Bool("pass", pass))
		return pass
	} else {
		log.Logger(ctx).Warn("There is a datasource filter but datasource name is not provided")
	}
	return true
}

// contextJobSameUuid checks if JobUuid can already be found in context and detects if it is the same
func (s *Subscriber) contextJobSameUuid(ctx context.Context, jobId string) bool {
	if mm, o := metadata.FromContextRead(ctx); o {
		if knownJob, ok := mm[strings.ToLower(servicecontext.ContextMetaJobUuid)]; ok && knownJob == jobId {
			return true
		}
		if knownJob, ok := mm[servicecontext.ContextMetaJobUuid]; ok && knownJob == jobId {
			return true
		}
	}
	return false
}

func createMessageFromEvent(event interface{}) *jobs.ActionMessage {
	initialInput := &jobs.ActionMessage{}

	if nodeChange, ok := event.(*tree.NodeChangeEvent); ok {
		ap, _ := anypb.New(nodeChange)
		initialInput.Event = ap
		if nodeChange.Target != nil {

			initialInput = initialInput.WithNode(nodeChange.Target)

		} else if nodeChange.Source != nil {

			initialInput = initialInput.WithNode(nodeChange.Source)

		}

	} else if triggerEvent, ok := event.(*jobs.JobTriggerEvent); ok {

		// Replace initialInput with TriggerMessage if it is set
		if triggerEvent.TriggerMessage != nil {
			initialInput = triggerEvent.TriggerMessage
			triggerEvent = proto.Clone(triggerEvent).(*jobs.JobTriggerEvent)
			triggerEvent.TriggerMessage = nil
		}
		initialInput.Event, _ = anypb.New(triggerEvent)

	} else if idmEvent, ok := event.(*idm.ChangeEvent); ok {

		ap, _ := anypb.New(idmEvent)
		initialInput.Event = ap
		if idmEvent.User != nil {
			initialInput = initialInput.WithUser(idmEvent.User)
		}
		if idmEvent.Role != nil {
			initialInput = initialInput.WithRole(idmEvent.Role)
		}
		if idmEvent.Workspace != nil {
			initialInput = initialInput.WithWorkspace(idmEvent.Workspace)
		}
		if idmEvent.Acl != nil {
			initialInput = initialInput.WithAcl(idmEvent.Acl)
		}

	}

	return initialInput
}

func logStartMessageFromEvent(ctx context.Context, event interface{}) {
	var msg string
	if triggerEvent, ok := event.(*jobs.JobTriggerEvent); ok {
		if triggerEvent.Schedule == nil {
			msg = "Starting job manually"
		} else {
			msg = "Starting job on schedule " + strings.ReplaceAll(triggerEvent.Schedule.String(), "Iso8601Schedule:", "")
		}
	} else if idmEvent, ok := event.(*idm.ChangeEvent); ok {
		eT := strings.ToLower(idmEvent.GetType().String())
		var oT string
		if idmEvent.User != nil {
			oT = "user"
		} else if idmEvent.Role != nil {
			oT = "role"
		} else if idmEvent.Workspace != nil {
			oT = "workspace"
		} else if idmEvent.Acl != nil {
			oT = "acl"
		}
		msg = fmt.Sprintf("Starting job on %s %s event", oT, eT)
	} else if nodeEvent, ok := event.(*tree.NodeChangeEvent); ok {
		eT := strings.ToLower(nodeEvent.GetType().String())
		msg = fmt.Sprintf("Starting job on %s node event", eT)
	}
	// Append user login
	user, _ := permissions.FindUserNameInContext(ctx)
	if user != "" && user != common.PydioSystemUsername {
		msg += " (triggered by user " + user + ")"
	}
	log.TasksLogger(ctx).Info(msg)
}

func (s *Subscriber) requiresUnsupportedCapacity(ctx context.Context, j *jobs.Job) error {
	if len(j.ResourcesDependencies) == 0 {
		return nil
	}
	for _, dep := range j.ResourcesDependencies {
		nodeDep := &rpb.Item{}
		if !dep.MessageIs(nodeDep) {
			continue
		}
		if e := dep.UnmarshalTo(nodeDep); e != nil {
			continue
		}
		if nodeDep.GetNode() == nil {
			continue
		}
		meta := nodeDep.GetMetadata()
		if meta == nil {
			continue
		}
		if err := runtime.MatchDependencies(ctx, meta); err != nil {
			log.Logger(ctx).Warn("Ignoring job "+j.Label+" as it requires unsupported capacities "+err.Error(), zap.Error(err))
			return err
		}
	}
	return nil
}
