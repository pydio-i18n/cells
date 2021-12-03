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

package service

import (
	"bufio"
	"context"
	"github.com/pydio/cells/v4/common/config/runtime"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/server/generic"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

// Fork
func Fork(f bool) ServiceOption {
	if f && !runtime.IsFork() {
		return WithGeneric(func(ctx context.Context, srv *generic.Server) error {
			runner := NewForkRunner(ctx)
			return runner.Start(ctx)
		})
	}

	return nil
}

// NewForkRunner creates a ChildrenRunner
func NewForkRunner(ctx context.Context) *ForkRunner {
	name := servicecontext.GetServiceName(ctx)
	c := &ForkRunner{
		name: name,
		initialCtx: ctx,
	}
	return c
}

// ChildrenRunner For Regexp based service
type ForkRunner struct {
	name              string
	initialCtx        context.Context
	cmd *exec.Cmd
}

// Start starts a forked process for a new source
func (f *ForkRunner) Start(ctx context.Context, retries ...int) error {

	f.cmd = exec.Command(os.Args[0], buildForkStartParams(f.name)...)

	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := f.cmd.StderrPipe()
	if err != nil {
		return err
	}

	serviceCtx := servicecontext.WithServiceName(ctx, f.name)
	scannerOut := bufio.NewScanner(stdout)
	go func() {
		for scannerOut.Scan() {
			text := strings.TrimRight(scannerOut.Text(), "\n")
			if strings.Contains(text, f.name) {
				log.StdOut.WriteString(text + "\n")
			} else {
				log.Logger(serviceCtx).Info(text)
			}
		}
	}()
	scannerErr := bufio.NewScanner(stderr)
	go func() {
		for scannerErr.Scan() {
			text := strings.TrimRight(scannerErr.Text(), "\n")
			if strings.Contains(text, f.name) {
				log.StdOut.WriteString(text + "\n")
			} else {
				log.Logger(serviceCtx).Error(text)
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	default:
		log.Logger(ctx).Debug("Starting SubProcess: " + f.name)
		if err := f.cmd.Start(); err != nil {
			return err
		}

		if err := f.cmd.Wait(); err != nil {
			if err.Error() != "signal: terminated" && err.Error() != "signal: interrupt" {
				log.Logger(serviceCtx).Error("SubProcess was not killed properly: " + err.Error())
				r := 0
				if len(retries) > 0 {
					r = retries[0]
				}
				if r < 3 {
					log.Logger(serviceCtx).Error("Restarting service in 3s...")
					<-time.After(3 * time.Second)
					return f.Start(ctx, r+1)
				}
			}
			return err
		}
	}

	return nil
}

// Stop services
func (f *ForkRunner) Stop(ctx context.Context) {
	log.Logger(ctx).Debug("stopping sub-process " + f.name)
	if f.cmd.Process != nil {
		if e := f.cmd.Process.Signal(syscall.SIGINT); e != nil {
			f.cmd.Process.Kill()
		}
	}
}
