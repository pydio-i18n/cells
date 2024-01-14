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
	"gocloud.dev/pubsub"
	"google.golang.org/protobuf/proto"
	"net/url"
	"strings"
	"sync"

	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/service/errors"
)

var (
	std = NewBroker("mem://")
)

func Register(b Broker) {
	std = b
}

func Default() Broker {
	return std
}

type Broker interface {
	PublishRaw(context.Context, string, []byte, map[string]string, ...PublishOption) error
	Publish(context.Context, string, proto.Message, ...PublishOption) error
	Subscribe(context.Context, string, SubscriberHandler, ...SubscribeOption) (UnSubscriber, error)
}

type UnSubscriber func() error

type SubscriberHandler func(ctx context.Context, msg Message) error

// NewBroker wraps a standard broker but prevents it from disconnecting while there still is a service running
func NewBroker(s string, opts ...Option) Broker {

	opts = append(opts, WithChainSubscriberInterceptor(
		TimeoutSubscriberInterceptor(),
		HeaderInjectorInterceptor(),
		// ContextInjectorInterceptor(),
	))

	options := newOptions(opts...)
	u, _ := url.Parse(s)
	scheme := u.Scheme

	br := &broker{
		publishOpener: func(ctx context.Context, topic string) (*pubsub.Topic, error) {
			namespace := u.Query().Get("namespace")
			uu := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: namespace + "/" + strings.TrimPrefix(topic, "/"), RawQuery: u.RawQuery}
			return pubsub.OpenTopic(ctx, uu.String())
		},
		subscribeOpener: func(topic string, oo ...SubscribeOption) (*pubsub.Subscription, error) {
			// Handle queue for grpc vs. nats vs memory
			op := &SubscribeOptions{Context: options.Context}
			for _, o := range oo {
				o(op)
			}

			q, _ := url.ParseQuery(u.RawQuery)
			ctx := op.Context
			if op.Queue != "" {
				switch scheme {
				case "nats", "grpc", "xds":
					q.Add("queue", op.Queue)
				default:
				}
			}

			namespace := u.Query().Get("namespace")

			uu := &url.URL{Scheme: u.Scheme, Host: u.Host, Path: namespace + "/" + strings.TrimPrefix(topic, "/"), RawQuery: q.Encode()}
			return pubsub.OpenSubscription(ctx, uu.String())
		},
		publishers: make(map[string]*pubsub.Topic),
		Options:    options,
	}

	chainPublisherInterceptors(br)
	chainSubscriberInterceptors(br)

	if options.Context != nil {
		go func() {
			<-options.Context.Done()
			br.closeTopics(options.Context)
		}()
	}

	return br
}

// PublishRaw sends a message to standard broker. For the moment, forward message to client.Publish
func PublishRaw(ctx context.Context, topic string, body []byte, header map[string]string, opts ...PublishOption) error {
	return std.PublishRaw(ctx, topic, body, header, opts...)
}

// Publish sends a message to standard broker. For the moment, forward message to client.Publish
func Publish(ctx context.Context, topic string, message proto.Message, opts ...PublishOption) error {
	return std.Publish(ctx, topic, message, opts...)
}

// MustPublish publishes a message ignoring the error
func MustPublish(ctx context.Context, topic string, message proto.Message, opts ...PublishOption) {
	err := Publish(ctx, topic, message)
	if err != nil {
		fmt.Printf("[Message Publication Error] Topic: %s, Error: %v\n", topic, err)
	}
}

func SubscribeCancellable(ctx context.Context, topic string, handler SubscriberHandler, opts ...SubscribeOption) error {
	// Go through Subscribe to parse MessageQueue option
	unsub, e := Subscribe(ctx, topic, handler, opts...)
	if e != nil {
		if errors.IsContextCanceled(e) {
			return nil
		}
		return e
	}
	go func() {
		<-ctx.Done()
		_ = unsub()
	}()

	return nil
}

func Subscribe(ctx context.Context, topic string, handler SubscriberHandler, opts ...SubscribeOption) (UnSubscriber, error) {
	so := parseSubscribeOptions(opts...)
	if so.MessageQueue != nil {
		qH := func(m Message) error {
			return so.MessageQueue.PushRaw(ctx, m)
		}
		er := so.MessageQueue.Consume(func(mm ...Message) {
			for _, m := range mm {
				if err := handler(m); err != nil {
					if so.ErrorHandler != nil {
						so.ErrorHandler(err)
					} else {
						fmt.Println("cannot apply message handler", err)
					}
				}
			}
		})
		if er != nil {
			return nil, er
		}
		// Replace original handler
		return std.Subscribe(ctx, topic, qH, opts...)
	}

	return std.Subscribe(ctx, topic, handler, opts...)
}

type broker struct {
	sync.Mutex
	publishOpener   TopicOpener
	subscribeOpener SubscribeOpener
	publishers      map[string]*pubsub.Topic
	Options
}

type TopicOpener func(context.Context, string) (*pubsub.Topic, error)
type SubscribeOpener func(string, ...SubscribeOption) (*pubsub.Subscription, error)

func (b *broker) openTopic(topic string) (*pubsub.Topic, error) {
	b.Lock()
	defer b.Unlock()
	publisher, ok := b.publishers[topic]
	if !ok {
		var err error
		publisher, err = b.publishOpener(b.Options.Context, topic)
		if err != nil {
			return nil, err
		}
		b.publishers[topic] = publisher
	}

	return publisher, nil
}

func (b *broker) closeTopics(c context.Context) {
	b.Lock()
	defer b.Unlock()
	for t, p := range b.publishers {
		_ = p.Shutdown(c)
		delete(b.publishers, t)
	}
}

func (b *broker) PublishRaw(ctx context.Context, topic string, body []byte, header map[string]string, opts ...PublishOption) error {
	publisher, err := b.openTopic(topic)
	if err != nil {
		return err
	}

	if err := publisher.Send(ctx, &pubsub.Message{
		Body:     body,
		Metadata: header,
	}); err != nil {
		return err
	}

	return nil
}

// Publish sends a message to standard broker. For the moment, forward message to client.Publish
func (b *broker) Publish(ctx context.Context, topic string, message proto.Message, opts ...PublishOption) error {
	body, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	header := make(map[string]string)
	if hh, ok := metadata.FromContextRead(ctx); ok {
		for k, v := range hh {
			header[k] = v
		}
	}

	publisher, err := b.openTopic(topic)
	if err != nil {
		return err
	}

	if err := publisher.Send(ctx, &pubsub.Message{
		Body:     body,
		Metadata: header,
	}); err != nil {
		return err
	}

	return nil
}

func (b *broker) Subscribe(ctx context.Context, topic string, handler SubscriberHandler, opts ...SubscribeOption) (UnSubscriber, error) {
	so := parseSubscribeOptions(opts...)

	// Making sure topic is opened
	_, err := b.openTopic(topic)
	if err != nil {
		return nil, err
	}

	sub, err := b.subscribeOpener(topic, opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			msg, err := sub.Receive(ctx)
			if err != nil {
				break
			}

			msg.Ack()

			if b.Options.subscriberInt != nil {
				if err := b.Options.subscriberInt(ctx, &message{
					header: msg.Metadata,
					body:   msg.Body,
				}, handler); err != nil {
					if so.ErrorHandler != nil {
						so.ErrorHandler(err)
					} else {
						fmt.Println("Cannot handle, no error handler set", topic, err.Error(), msg.Metadata, string(msg.Body))
					}
				}
			} else {
				if err := handler(ctx, &message{
					header: msg.Metadata,
					body:   msg.Body,
				}); err != nil {
					if so.ErrorHandler != nil {
						so.ErrorHandler(err)
					} else {
						fmt.Println("Cannot handle, no error handler set", topic, err.Error(), msg.Metadata, string(msg.Body))
					}
				}
			}
		}
	}()

	return func() error {
		return sub.Shutdown(ctx)
	}, nil
}
