package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common"
	cgrpc "github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/jobs"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/metadata"
)

type ReconnectingClient struct {
	parentCtx context.Context
	stopChan  chan bool
	closed    bool
}

type TaskStatusUpdate struct {
	*jobs.Task
	RunnableContext context.Context
	RunnableStatus  jobs.TaskStatus
}

func NewTaskReconnectingClient(parentCtx context.Context) *ReconnectingClient {
	r := &ReconnectingClient{
		parentCtx: parentCtx,
		stopChan:  make(chan bool),
	}
	return r
}

func (s *ReconnectingClient) StartListening(tasksChan chan interface{}) {
	s.chanToStream(tasksChan)
}

func (s *ReconnectingClient) Stop() {
	s.stopChan <- true
}

func (s *ReconnectingClient) chanToStream(ch chan interface{}) {

	go func() {
		taskClient := jobs.NewJobServiceClient(cgrpc.GetClientConnFromCtx(s.parentCtx, common.ServiceJobs))
		ct, ca := context.WithCancel(s.parentCtx)
		defer ca()
		streamer, e := taskClient.PutTaskStream(ct)
		if e != nil {
			log.Logger(s.parentCtx).Error("Streamer PutTaskStream", zap.Error(e))
			ca()
			<-time.After(10 * time.Second)
			s.chanToStream(ch)
			return
		}

		postAndRetry := func(er error, req *jobs.PutTaskRequest) {
			log.Logger(s.parentCtx).Debug("Cannot post task - break and reconnect streamer", zap.Error(er))
			_ = streamer.CloseSend()
			ca()
			if _, rE := taskClient.PutTask(s.parentCtx, req); rE == nil {
				log.Logger(s.parentCtx).Debug("Posted with a direct request")
			}
			if !s.closed {
				<-time.After(1 * time.Second)
				s.chanToStream(ch)
			}
		}

		for {
			select {
			case val := <-ch:
				var request *jobs.PutTaskRequest
				if t, ok := val.(*jobs.Task); ok {
					request = &jobs.PutTaskRequest{Task: t}
				} else if ts, ok2 := val.(*TaskStatusUpdate); ok2 {
					request = &jobs.PutTaskRequest{
						Task: ts.Task,
					}
					if ts.RunnableContext != nil {
						request.StatusMeta = map[string]string{}
						if mm, has := metadata.FromContextRead(ts.RunnableContext); has {
							if p, o := mm[servicecontext.ContextMetaTaskActionPath]; o {
								request.StatusMeta[servicecontext.ContextMetaTaskActionPath] = p
							}
							if tags, o := mm[servicecontext.ContextMetaTaskActionTags]; o {
								request.StatusMeta[servicecontext.ContextMetaTaskActionTags] = tags
							}
						}
						request.StatusMeta["X-Pydio-Task-Action-Status"] = ts.RunnableStatus.String()
					}
				} else if val != nil {
					log.Logger(s.parentCtx).Error("Unexpected: could not cast value to jobs.Task", zap.Any("val", val))
				}
				if request != nil {
					if se := streamer.Send(request); se != nil {
						postAndRetry(se, request)
						return
					}
					if _, re := streamer.Recv(); re != nil {
						postAndRetry(re, request)
						return
					}
				}
			case <-s.stopChan:
				s.closed = true
				return
			}
		}
	}()

}
