/*
 * Copyright (c) 2024. Abstrium SAS <team (at) pydio.com>
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

package debounce

import (
	"context"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v5/common/broker"
	"github.com/pydio/cells/v5/common/errors"
	"github.com/pydio/cells/v5/common/runtime"
	"github.com/pydio/cells/v5/common/runtime/controller"
	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/common/utils/propagator"
)

// protoWithContext composes a generic type and a context
type protoWithContext struct {
	Original proto.Message
	Ctx      context.Context
}

func (t *protoWithContext) Unmarshal(ctx context.Context, target proto.Message) (context.Context, error) {
	// We assume that expected output will be of the same type
	originalValue := reflect.ValueOf(t.Original).Elem()
	targetValue := reflect.ValueOf(target).Elem()

	// If they're not of the same type, return an error
	if originalValue.Type() != targetValue.Type() {
		return ctx, errors.New("t.Original and target are not of the same type")
	}

	// Copy the fields from t.Original to target
	targetValue.Set(originalValue)
	return ctx, nil
}

func (t *protoWithContext) RawData() (map[string]string, []byte) {
	return map[string]string{}, []byte{}
}

func init() {
	runtime.Register("system", func(ctx context.Context) {
		var mgr manager.Manager
		if !propagator.Get(ctx, manager.ContextKey, &mgr) {
			return
		}
		mgr.RegisterQueue("mem", controller.WithCustomOpener(func(ctx context.Context, url string) (broker.AsyncQueuePool, error) {
			return broker.NewWrappedPool(url, broker.MakeWrappedOpener(&debounce{}))
		}))
	})

}

// debounce debounces events on a given timeframe and calls process on them afterward
// Use globalCtx.Done() to stop listening to events
type debounce struct {
	Events chan broker.Message
	batch  []broker.Message

	Done            chan bool
	globalCtx       context.Context
	cancel          context.CancelFunc
	debounce        time.Duration
	idle            time.Duration
	maxEvents       int
	processCallback func(ctx context.Context, message ...broker.Message)
	closed          bool
}

// OpenURL creates a new *debounce{} to be used as broker.AsyncQueue
func (b *debounce) OpenURL(ctx context.Context, u *url.URL) (broker.AsyncQueue, error) {
	deb := 3 * time.Second
	idl := 20 * time.Second
	maxN := 2000

	if d := u.Query().Get("debounce"); d != "" {
		if du, er := time.ParseDuration(d); er == nil {
			deb = du
		} else {
			log.Logger(ctx).Warn("[debounce] Cannot parse debounce, using default", zap.Error(er))
		}
	}
	if d := u.Query().Get("idle"); d != "" {
		if du, er := time.ParseDuration(d); er == nil {
			idl = du
		} else {
			log.Logger(ctx).Warn("[debounce] Cannot parse idle, using default", zap.Error(er))
		}
	}

	if m := u.Query().Get("max"); m != "" {
		if ma, er := strconv.Atoi(m); er == nil {
			maxN = ma
		} else {
			log.Logger(ctx).Warn("[debounce] Cannot parse max, using default", zap.Error(er))
		}
	}

	return &debounce{
		Events:    make(chan broker.Message, 1000),
		globalCtx: ctx,
		debounce:  deb,
		idle:      idl,
		maxEvents: maxN,
	}, nil
}

// Consume registers the processor as callback and starts listening to queue
func (b *debounce) Consume(process func(context.Context, ...broker.Message)) error {
	b.processCallback = process
	go b.Start()
	return nil
}

// Start starts listening to incoming events
func (b *debounce) Start() {
	next := b.debounce
	defer func() {
		b.closed = true
		close(b.Events)
	}()
	ctx, can := context.WithCancel(b.globalCtx)
	b.cancel = can
	for {
		select {
		case e := <-b.Events:
			b.batch = append(b.batch, e)
			if b.maxEvents > 0 && len(b.batch) > b.maxEvents {
				b.process()
			}
			next = b.debounce
		case <-time.After(next):
			b.process()
			next = b.idle
		case <-ctx.Done():
			b.process()
			return
		}
	}
}

// Push sends a message to the queue
func (b *debounce) Push(ctx context.Context, msg proto.Message) error {
	if b.closed {
		return errors.New("channel is already closed")
	}
	b.Events <- &protoWithContext{Original: msg, Ctx: ctx}
	return nil
}

func (b *debounce) PushRaw(_ context.Context, message broker.Message) error {
	if b.closed {
		return errors.New("channel is already closed")
	}
	b.Events <- message
	return nil
}

func (b *debounce) process() {
	if len(b.batch) == 0 {
		return
	}
	go b.processBatch(b.batch)
	b.batch = []broker.Message{}
}

func (b *debounce) processBatch(bb []broker.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Logger(b.globalCtx).Error("Recovered in debouncer", zap.Any("r", r), zap.Stack("stack"))
		}
	}()

	log.Logger(b.globalCtx).Debug("Processing batched events", zap.Int("size", len(bb)))
	var cleanEvents []broker.Message
	for _, e := range bb {
		cleanEvents = append(cleanEvents, e)
	}
	b.processCallback(b.globalCtx, cleanEvents...)
}

func (b *debounce) Close(ctx context.Context) error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}
