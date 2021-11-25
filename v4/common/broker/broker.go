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
	"fmt"

	"github.com/pydio/cells/v4/common/service/context/metadata"

	"google.golang.org/protobuf/proto"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/broker/memory"
)

type brokerwrap struct {
	b    broker.Broker
	opts Options
}

var (
	std = NewBroker(memory.NewBroker())
)

// NewBroker wraps a standard broker but prevents it from disconnecting while there still is a service running
func NewBroker(b broker.Broker, opts ...Option) broker.Broker {
	return &brokerwrap{b, newOptions(opts...)}
}

func Connect() error {
	return std.Connect()
}

func Disconnect() error {
	return std.Disconnect()
}

// Publish sends a message to standard broker. For the moment, forward message to client.Publish
func Publish(ctx context.Context, topic string, message interface{}, opts ...PublishOption) error {
	body, _ := proto.Marshal(message.(proto.Message))
	header := make(map[string]string)
	if hh, ok := metadata.FromContext(ctx); ok {
		for k, v := range hh {
			header[k] = v
		}
	}
	return std.Publish(topic, &broker.Message{Body: body, Header: header}, broker.PublishContext(ctx))
}

// MustPublish publishes a message ignoring the error
func MustPublish(ctx context.Context, topic string, message interface{}, opts ...PublishOption) {
	err := Publish(ctx, topic, message)
	if err != nil {
		fmt.Printf("[Message Publication Error] Topic: %s, Error: %v\n", topic, err)
	}
}

func SubscribeCancellable(ctx context.Context, topic string, handler SubscriberHandler, opts ...SubscribeOption) error {
	unsub, e := Subscribe(topic, handler, opts...)
	if e != nil {
		return e
	}
	go func() {
		<-ctx.Done()
		_ = unsub()
	}()
	return nil
}

func Subscribe(topic string, handler SubscriberHandler, opts ...SubscribeOption) (UnSubscriber, error) {

	so := &SubscribeOptions{}
	for _, o := range opts {
		o(so)
	}
	var mopts []broker.SubscribeOption
	if so.Context != nil {
		mopts = append(mopts, broker.SubscribeContext(so.Context))
	}
	if so.Queue != "" {
		mopts = append(mopts, broker.Queue(so.Queue))
	}
	if so.ErrorHandler != nil {
		mopts = append(mopts, broker.HandleError(func(message *broker.Message, err error) {
			so.ErrorHandler(err)
		}))
	}

	sub, er := std.Subscribe(topic, func(message *broker.Message) error {
		subMsg := &SubMessage{Message: message}
		return handler(subMsg)
	}, mopts...)
	if er != nil {
		return nil, er
	}
	return sub.Unsubscribe, nil

}

/*
func Subscribe(s string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	return std.Subscribe(s, h, opts...)
}
*/

// Options wraps standard function
func (b *brokerwrap) Options() broker.Options {
	return b.b.Options()
}

// Address wraps standard function
func (b *brokerwrap) Address() string {
	return b.b.Address()
}

// Connect wraps standard function
func (b *brokerwrap) Connect() error {
	return b.b.Connect()
}

// Disconnect handles the disconnection to the broker. It prevents it if there is a service that is still active
func (b *brokerwrap) Disconnect() error {
	for _, o := range b.opts.beforeDisconnect {
		if err := o(); err != nil {
			return err
		}
	}

	return b.b.Disconnect()
}

// Init wraps standard function
func (b *brokerwrap) Init(opts ...broker.Option) error {
	return b.b.Init(opts...)
}

// Publish wraps standard function
func (b *brokerwrap) Publish(s string, m *broker.Message, opts ...broker.PublishOption) error {
	return b.b.Publish(s, m, opts...)
}

// Publish wraps standard function
func (b *brokerwrap) Subscribe(s string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	return b.b.Subscribe(s, h, opts...)
}

// Publish wraps standard function
func (b *brokerwrap) String() string {
	return b.b.String()
}
