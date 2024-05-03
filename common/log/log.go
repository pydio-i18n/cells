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

package log

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/runtime/runtimecontext"
)

// WriteSyncer implements zapcore.WriteSyncer
type WriteSyncer interface {
	io.Writer
	Sync() error
}

type LogContextWrapper func(ctx context.Context, logger ZapLogger, fields ...zapcore.Field) ZapLogger

var (
	mainLogger     = newLogger()
	auditLogger    = newLogger()
	tasksLogger    = newLogger()
	contextWrapper = BasicContextWrapper

	mainLogSyncerClient *LogSyncer

	StdOut *os.File

	skipServerSync bool
	customSyncers  []zapcore.WriteSyncer
	// Parse log lines like below:
	// 2022/04/21 07:53:46.226	[33mWARN[0m	tls	stapling OCSP	{"error": "no OCSP stapling for [charles-pydio.local kubernetes.docker.internal local.pydio local.pydio.com localhost localhost localpydio.com spnego.lab.py sub1.pydio sub2.pydio 127.0.0.1 192.168.10.101]: no OCSP server specified in certificate"}
	caddyInternals = regexp.MustCompile("^(?P<log_date>[^\t]+)\t(?P<log_level>[^\t]+)\t(?P<log_name>[^\t]+)\t(?P<log_message>[^\t]+)(\t)?(?P<log_fields>[^\t]+)$")
)

// Init for the log package - called by the main
func Init(logDir string, ww ...LogContextWrapper) {
	SetLoggerInit(func() *zap.Logger {

		StdOut = os.Stdout

		var logger *zap.Logger

		serverCore := zapcore.NewNopCore()
		if !skipServerSync && mainLogSyncerClient != nil {
			// Create core for internal indexing service
			// It forwards logs to the pydio.grpc.logs service to store them
			// Format is always JSON + ProductionEncoderConfig
			srvConfig := zap.NewProductionEncoderConfig()
			srvConfig.EncodeTime = RFC3369TimeEncoder
			serverSync := zapcore.AddSync(mainLogSyncerClient)
			serverCore = zapcore.NewCore(
				zapcore.NewJSONEncoder(srvConfig),
				serverSync,
				getCoreLevel(),
			)
		}

		syncers := []zapcore.WriteSyncer{StdOut}
		if common.LogToFile {
			// Additional logger: stores messages in local file
			// logDir := config.ApplicationWorkingDir(config.ApplicationDirLogs)
			rotaterSync := zapcore.AddSync(&lumberjack.Logger{
				Filename:   filepath.Join(logDir, "pydio.log"),
				MaxSize:    10, // megabytes
				MaxBackups: 100,
				MaxAge:     28, // days
			})
			syncers = append(syncers, rotaterSync)
		}
		syncers = append(syncers, customSyncers...)
		syncer := zapcore.NewMultiWriteSyncer(syncers...)

		if common.LogConfig == common.LogConfigProduction {

			// lumberjack.Logger is already safe for concurrent use, so we don't need to lock it.
			cfg := zap.NewProductionEncoderConfig()
			cfg.EncodeTime = RFC3369TimeEncoder
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg),
				syncer,
				getCoreLevel(),
			)
			core = zapcore.NewTee(core, serverCore)
			logger = zap.New(core)

		} else {

			cfg := zap.NewDevelopmentEncoderConfig()
			cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			core := zapcore.NewCore(
				newColorConsoleEncoder(cfg),
				syncer,
				getCoreLevel(),
			)
			core = zapcore.NewTee(core, serverCore)
			if getCoreLevel() == zap.DebugLevel || traceFatalEnabled() {
				logger = zap.New(core, zap.AddStacktrace(zap.ErrorLevel))
			} else {
				logger = zap.New(core)
			}

		}

		if traceFatalEnabled() {
			_, _ = zap.RedirectStdLogAt(logger, zap.ErrorLevel) // log anything at ErrorLevel with a stack trace
		} else {
			_, _ = zap.RedirectStdLogAt(logger, zap.DebugLevel)
		}

		return logger
	}, func(ctx context.Context) {
		if !skipServerSync {
			mainLogSyncerClient = NewLogSyncer(ctx, common.ServiceLog)
		}
	})
	if len(ww) > 0 {
		contextWrapper = ww[0]
	}
}

func CaptureCaddyStdErr(serviceName string) context.Context {
	ctx := runtimecontext.WithServiceName(context.Background(), serviceName)
	lg := Logger(ctx)
	if traceFatalEnabled() {
		return ctx
	}
	r, w, err := os.Pipe()
	if err != nil {
		return ctx
	}
	rootErr := os.Stderr
	os.Stderr = w
	//caddyLogger := logger.Named("pydio.server.caddy")
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if parsed := caddyInternals.FindStringSubmatch(line); len(parsed) == 7 {
				level := strings.Trim(parsed[2], "[340m ")
				msg := parsed[3] + " - " + parsed[4] + parsed[6]
				if strings.Contains(level, "INFO") {
					lg.Info(msg)
				} else if strings.Contains(level, "WARN") {
					lg.Warn(msg)
				} else if strings.Contains(level, "ERROR") {
					lg.Error(msg)
				} else { // DEBUG, WARN, or other value
					lg.Debug(msg)
				}
			} else {
				_, _ = rootErr.WriteString(line)
			}
		}
	}()
	return ctx
}

// RegisterWriteSyncer optional writers for logs
func RegisterWriteSyncer(syncer WriteSyncer) {
	customSyncers = append(customSyncers, syncer)

	mainLogger.forceReset() // Will force reinit next time
}

// SetSkipServerSync can disable the core syncing to cells service
// Must be called before initialization
func SetSkipServerSync() {
	skipServerSync = true
}

// initLogger sets the initializer and eventually registers a GlobalConnConsumer function.
func initLogger(l *logger, f func() *zap.Logger, globalConnInit func(ctx context.Context)) {
	l.set(f)
	if globalConnInit != nil {
		runtime.RegisterGlobalConnConsumer("main", func(ctx context.Context) {
			globalConnInit(ctx)
			l.forceReset()
		})
	}
}

// SetLoggerInit defines what function to use to init the logger
func SetLoggerInit(f func() *zap.Logger, globalConnInit func(ctx context.Context)) {
	initLogger(mainLogger, f, globalConnInit)
}

// Logger returns a zap logger with as much context as possible.
func Logger(ctx context.Context) ZapLogger {
	/*
		// Todo recheck - WithLogger was never used anywhere
		l := runtimecontext.GetLogger(ctx)
		if l != nil {
			if lg, ok := l.(ZapLogger); ok {
				return lg
			}
		}
	*/
	return contextWrapper(ctx, mainLogger.get())
}

// SetAuditerInit defines what function to use to init the auditer
func SetAuditerInit(f func() *zap.Logger, globalConnInit func(ctx context.Context)) {
	initLogger(auditLogger, f, globalConnInit)
}

// Auditer returns a zap logger with as much context as possible
func Auditer(ctx context.Context) ZapLogger {
	return contextWrapper(ctx, auditLogger.get(), zap.String("LogType", "audit"))
}

// SetTasksLoggerInit defines what function to use to init the tasks logger
func SetTasksLoggerInit(f func() *zap.Logger, globalConnInit func(ctx context.Context)) {
	initLogger(tasksLogger, f, globalConnInit)
}

// TasksLogger returns a zap logger with as much context as possible.
func TasksLogger(ctx context.Context) ZapLogger {
	return contextWrapper(ctx, tasksLogger.get(), zap.String("LogType", "tasks"))
}

// GetAuditId simply returns a zap field that contains this message id to ease audit log analysis.
func GetAuditId(msgId string) zapcore.Field {
	return zap.String(common.KeyMsgId, msgId)
}

// RFC3369TimeEncoder serializes a time.Time to an RFC3339-formatted string
func RFC3369TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

func Debug(msg string, fields ...zapcore.Field) {
	mainLogger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	mainLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	mainLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	mainLogger.Fatal(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	mainLogger.Info(msg, fields...)
}
