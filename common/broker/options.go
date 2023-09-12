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

package broker

import (
	"context"

	"github.com/pydio/cells/v4/common/utils/uuid"
)

// Options to the broker
type Options struct {
	Context          context.Context
	beforeDisconnect []func() error
}

// Option definition
type Option func(*Options)

func newOptions(opts ...Option) Options {
	opt := Options{}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func WithContext(ctx context.Context) Option {
	return func(options *Options) {
		options.Context = ctx
	}
}

// BeforeDisconnect registers all functions to be triggered before the broker disconnect
func BeforeDisconnect(f func() error) Option {
	return func(o *Options) {
		o.beforeDisconnect = append(o.beforeDisconnect, f)
	}
}

type PublishOptions struct {
	Context context.Context
}
type PublishOption func(options *PublishOptions)

func PublishContext(ctx context.Context) PublishOption {
	return func(options *PublishOptions) {
		options.Context = ctx
	}
}

type SubscribeOptions struct {
	// Handler executed when errors occur processing messages
	ErrorHandler func(error)

	// Subscribers with the same queue name
	// will create a shared subscription where each
	// receives a subset of messages.
	Queue string

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context

	// Optional MessageQueue than can debounce/persist
	// received messages and re-process them later on
	MessageQueue MessageQueue

	// Optional name for metrics
	CounterName string
}

type SubscribeOption func(*SubscribeOptions)

func parseSubscribeOptions(opts ...SubscribeOption) SubscribeOptions {
	opt := SubscribeOptions{
		CounterName: uuid.New()[0:6],
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// HandleError sets an ErrorHandler to catch all broker errors that cant be handled
// in normal way, for example Codec errors
func HandleError(h func(error)) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.ErrorHandler = h
	}
}

// Queue sets the name of the queue to share messages on
func Queue(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Queue = name
	}
}

// WithLocalQueue passes a FIFO queue to absorb input
func WithLocalQueue(q MessageQueue) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.MessageQueue = q
	}
}

// WithCounterName adds a custom id for metrics counter name
func WithCounterName(n string) SubscribeOption {
	return func(options *SubscribeOptions) {
		options.CounterName = n
	}
}

// SubscribeContext set context
func SubscribeContext(ctx context.Context) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Context = ctx
	}
}
