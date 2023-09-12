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

// Package grpc provides a gRPC service to access the store for scheduler job definitions.
package grpc

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/broker/log"
	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	grpc2 "github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/dao/bleve"
	"github.com/pydio/cells/v4/common/dao/boltdb"
	"github.com/pydio/cells/v4/common/dao/mongodb"
	log3 "github.com/pydio/cells/v4/common/log"
	proto "github.com/pydio/cells/v4/common/proto/jobs"
	log2 "github.com/pydio/cells/v4/common/proto/log"
	"github.com/pydio/cells/v4/common/proto/sync"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/scheduler/jobs"
)

var (
	Migration140 = false
	Migration150 = false
	Migration230 = false
)

const ServiceName = common.ServiceGrpcNamespace_ + common.ServiceJobs

func init() {
	defaults := getDefaultJobs()
	for _, j := range defaults {
		proto.RegisterDefault(j, ServiceName)
	}

	runtime.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(ServiceName),
			service.Context(ctx),
			service.Tag(common.ServiceTagScheduler),
			service.Description("Store for scheduler jobs description"),
			// service.Unique(true),
			service.Fork(true),
			service.WithStorage(jobs.NewDAO,
				service.WithStoragePrefix("jobs"),
				service.WithStorageMigrator(jobs.Migrate),
				service.WithStorageSupport(boltdb.Driver, mongodb.Driver),
				service.WithStorageDefaultDriver(func() (string, string) {
					return boltdb.Driver, filepath.Join(runtime.MustServiceDataDir(ServiceName), "jobs.db")
				}),
			),
			service.WithIndexer(log.NewDAO,
				service.WithStoragePrefix("tasklogs"),
				service.WithStorageMigrator(log.Migrate),
				service.WithStorageSupport(bleve.Driver, mongodb.Driver),
				service.WithStorageDefaultDriver(func() (string, string) {
					return bleve.Driver, filepath.Join(runtime.MustServiceDataDir(ServiceName), "tasklogs.bleve?mapping=log&rotationSize=-1")
				}),
			),
			service.Migrations([]*service.Migration{
				{
					TargetVersion: service.ValidVersion("1.4.0"),
					Up: func(ctx context.Context) error {
						// Set flag for migration script to be run AfterStart (see below, handler cannot be shared)
						Migration140 = true
						return nil
					},
				},
				{
					TargetVersion: service.ValidVersion("1.5.0"),
					Up: func(ctx context.Context) error {
						// Set flag for migration script to be run AfterStart (see below, handler cannot be shared)
						Migration150 = true
						return nil
					},
				},
				{
					TargetVersion: service.ValidVersion("2.2.99"),
					Up: func(ctx context.Context) error {
						// Set flag for migration script to be run AfterStart (see below, handler cannot be shared)
						Migration230 = true
						return nil
					},
				},
			}),
			service.WithGRPC(func(c context.Context, server grpc.ServiceRegistrar) error {

				store := servicecontext.GetDAO(c).(jobs.DAO)
				index := servicecontext.GetIndexer(c).(dao.IndexDAO)

				logStore, err := log.NewIndexService(index)
				if err != nil {
					return err
				}
				handler := NewJobsHandler(c, store, logStore)
				proto.RegisterJobServiceEnhancedServer(server, handler)
				log2.RegisterLogRecorderEnhancedServer(server, handler)
				sync.RegisterSyncEndpointEnhancedServer(server, handler)
				logger := log3.Logger(c)

				for _, j := range defaults {
					if _, e := handler.GetJob(c, &proto.GetJobRequest{JobID: j.ID}); e != nil {
						_, _ = handler.PutJob(c, &proto.PutJobRequest{Job: j})
					}
					// Force re-adding thumbs job
					if Migration230 && j.ID == "thumbs-job" {
						_, _ = handler.PutJob(c, &proto.PutJobRequest{Job: j})
					}
				}
				// Clean tasks stuck in "Running" status
				if _, er := handler.CleanStuckTasks(c, true, logger); er != nil {
					logger.Warn("Could not run CleanStuckTasks: "+er.Error(), zap.Error(er))
				}

				// Clean user-jobs (AutoStart+AutoClean) without any tasks
				if er := handler.CleanDeadUserJobs(c); er != nil {
					logger.Warn("Could not run CleanDeadUserJobs: "+er.Error(), zap.Error(er))
				}

				if Migration140 {
					if resp, e := handler.DeleteTasks(c, &proto.DeleteTasksRequest{
						JobId:      "users-activity-digest",
						Status:     []proto.TaskStatus{proto.TaskStatus_Any},
						PruneLimit: 1,
					}); e == nil {
						logger.Info("Migration 1.4.0: removed tasks on job users-activity-digest that could fill up the scheduler", zap.Int("number", len(resp.Deleted)))
					} else {
						logger.Error("Error while trying to prune tasks for job users-activity-digest", zap.Error(e))
					}
					if resp, e := handler.DeleteTasks(c, &proto.DeleteTasksRequest{
						JobId:      "resync-changes-job",
						Status:     []proto.TaskStatus{proto.TaskStatus_Any},
						PruneLimit: 1,
					}); e == nil {
						logger.Info("Migration 1.4.0: removed tasks on job resync-changes-job that could fill up the scheduler", zap.Int("number", len(resp.Deleted)))
					} else {
						logger.Error("Error while trying to prune tasks for job resync-changes-job", zap.Error(e))
					}
				}
				if Migration150 {
					// Remove archive-changes-job
					if _, e := handler.DeleteJob(c, &proto.DeleteJobRequest{JobID: "archive-changes-job"}); e != nil {
						logger.Error("Could not remove archive-changes-job", zap.Error(e))
					} else {
						logger.Info("[Migration] Removed archive-changes-job")
					}
					// Remove resync-changes-job
					if _, e := handler.DeleteJob(c, &proto.DeleteJobRequest{JobID: "resync-changes-job"}); e != nil {
						logger.Error("Could not remove resync-changes-job", zap.Error(e))
					} else {
						logger.Info("[Migration] Removed resync-changes-job")
					}
				}
				if Migration230 {
					// Remove clean thumbs job and re-insert thumbs job
					if _, e := handler.DeleteJob(c, &proto.DeleteJobRequest{JobID: "clean-thumbs-job"}); e != nil {
						logger.Error("Could not remove clean-thumbs-job", zap.Error(e))
					} else {
						logger.Info("[Migration] Removed clean-thumbs-job")
					}
				}

				var hc grpc2.HealthMonitor
				if jj, e := handler.ListAutoRestartJobs(ctx); e == nil && len(jj) > 0 {
					// We should wait for service task to be started, then start jobs
					hc = grpc2.NewHealthChecker(ctx)
					go func() {
						hc.Monitor(common.ServiceTasks)
						for _, j := range jj {
							logger.Info("Sending a start event for job '" + j.Label + "'")
							_ = broker.Publish(c, common.TopicTimerEvent, &proto.JobTriggerEvent{
								JobID:  j.ID,
								RunNow: true,
							})
						}
					}()
				}

				go func() {
					<-c.Done()
					handler.Close()
					logStore.Close(c)
					if hc != nil {
						hc.Stop()
					}
				}()
				return nil
			}),
		)

	})
}
