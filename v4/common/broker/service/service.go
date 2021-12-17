// Copyright 2018 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/pydio/cells/v4/common"
	grpc2 "github.com/pydio/cells/v4/common/client/grpc"
	pb "github.com/pydio/cells/v4/common/proto/broker"
	"gocloud.dev/gcerrors"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/driver"
	"google.golang.org/grpc"
)

func init() {
	o := new(URLOpener)
	pubsub.DefaultURLMux().RegisterTopic(Scheme, o)
	pubsub.DefaultURLMux().RegisterSubscription(Scheme, o)
}

// Scheme is the URL scheme grpc pubsub registers its URLOpeners under on pubsub.DefaultMux.
const Scheme = "grpc"

// URLOpener opens grpc pubsub URLs like "cells://topic".
//
// The URL's host+path is used as the topic to create or subscribe to.
//
// Query parameters:
//   - ackdeadline: The ack deadline for OpenSubscription, in time.ParseDuration formats.
//       Defaults to 1m.
type URLOpener struct {
	mu     sync.Mutex
	topics map[string]*pubsub.Topic
}

// OpenTopicURL opens a pubsub.Topic based on u.
func (o *URLOpener) OpenTopicURL(ctx context.Context, u *url.URL) (*pubsub.Topic, error) {
	topicName := u.Path
	return NewTopic(topicName)
}

// OpenSubscriptionURL opens a pubsub.Subscription based on u.
func (o *URLOpener) OpenSubscriptionURL(ctx context.Context, u *url.URL) (*pubsub.Subscription, error) {
	//q := u.Query()
	//
	//ackDeadline := 1 * time.Minute
	//if s := q.Get("ackdeadline"); s != "" {
	//	var err error
	//	ackDeadline, err = time.ParseDuration(s)
	//	if err != nil {
	//		return nil, fmt.Errorf("open subscription %v: invalid ackdeadline %q: %v", u, s, err)
	//	}
	//	q.Del("ackdeadline")
	//}
	//for param := range q {
	//	return nil, fmt.Errorf("open subscription %v: invalid query parameter %q", u, param)
	//}

	topicName := u.Path
	return NewSubscription(topicName)
}

var errNotExist = errors.New("cellspubsub: topic does not exist")

type topic struct {
	path   string
	stream pb.Broker_PublishClient
}

// NewTopic creates a new in-memory topic.
func NewTopic(path string, opts ...Option) (*pubsub.Topic, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// extract the client from the context, fallback to grpc
	var conn grpc.ClientConnInterface
	if ctx != nil {
		if v := ctx.Value(clientKey{}); v != nil {
			if c, ok := v.(grpc.ClientConnInterface); ok {
				conn = c
			}
		}
	}

	if conn == nil {
		conn = grpc2.NewClientConn(common.ServiceBroker)
	}

	fmt.Println("Create topic later")

	cli := pb.NewBrokerClient(conn)
	stream, err := cli.Publish(ctx)
	if err != nil {
		fmt.Println("And we have an error ? ", err)
		return nil, err
	}
	fmt.Println("And the stream is initialized ", stream)

	return pubsub.NewTopic(&topic{
		path:   path,
		stream: stream,
	}, nil), nil
}

// SendBatch implements driver.Topic.SendBatch.
// It is error if the topic is closed or has no subscriptions.
func (t *topic) SendBatch(ctx context.Context, dms []*driver.Message) error {
	if t == nil || t.stream == nil {
		return errors.New("nil variable")
	}

	var ms []*pb.Message
	for _, dm := range dms {
		psm := &pb.Message{Body: dm.Body, Header: dm.Metadata}
		if dm.BeforeSend != nil {
			asFunc := func(i interface{}) bool {
				if p, ok := i.(**pb.Message); ok {
					*p = psm
					return true
				}
				return false
			}
			if err := dm.BeforeSend(asFunc); err != nil {
				return err
			}
		}
		ms = append(ms, psm)
	}
	req := &pb.PublishRequest{Topic: t.path, Messages: &pb.Messages{Messages: ms}}
	if err := t.stream.Send(req); err != nil {
		return err
	}

	//for n, dm := range dms {
	//	if dm.AfterSend != nil {
	//		asFunc := func(i interface{}) bool {
	//			if p, ok := i.(*string); ok {
	//				*p = pr.MessageIds[n]
	//				return true
	//			}
	//			return false
	//		}
	//		if err := dm.AfterSend(asFunc); err != nil {
	//			return err
	//		}
	//	}
	//}

	return nil
}

// IsRetryable implements driver.Topic.IsRetryable.
func (*topic) IsRetryable(error) bool { return false }

// As implements driver.Topic.As.
// It supports *topic so that NewSubscription can recover a *topic
// from the portable type (see below). External users won't be able
// to use As because topic isn't exported.
func (t *topic) As(i interface{}) bool {
	x, ok := i.(**topic)
	if !ok {
		return false
	}
	*x = t
	return true
}

// ErrorAs implements driver.Topic.ErrorAs
func (*topic) ErrorAs(error, interface{}) bool {
	return false
}

// ErrorCode implements driver.Topic.ErrorCode
func (*topic) ErrorCode(err error) gcerrors.ErrorCode {
	if err == errNotExist {
		return gcerrors.NotFound
	}
	return gcerrors.Unknown
}

// Close implements driver.Topic.Close.
func (*topic) Close() error { return nil }

type subscription struct {
	cli pb.Broker_SubscribeClient
}

// NewSubscription returns a *pubsub.Subscription representing a NATS subscription or NATS queue subscription.
// The subject is the NATS Subject to subscribe to;
// for more info, see https://nats.io/documentation/writing_applications/subjects.
func NewSubscription(path string, opts ...Option) (*pubsub.Subscription, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// extract the client from the context, fallback to grpc
	var conn grpc.ClientConnInterface
	if ctx != nil {
		if v := ctx.Value(clientKey{}); v != nil {
			if c, ok := v.(grpc.ClientConnInterface); ok {
				conn = c
			}
		}
	}

	if conn == nil {
		conn = grpc2.NewClientConn(common.ServiceBroker)
	}

	req := &pb.SubscribeRequest{Topic: path, Queue: options.Queue}
	cli, err := pb.NewBrokerClient(conn).Subscribe(ctx, req)
	if err != nil {
		return nil, err
	}

	fmt.Println("Created new subscription")
	return pubsub.NewSubscription(&subscription{
		cli: cli,
	}, nil, nil), nil
}

// ReceiveBatch implements driver.ReceiveBatch.
func (s *subscription) ReceiveBatch(ctx context.Context, maxMessages int) ([]*driver.Message, error) {
	if s == nil || s.cli == nil {
		return nil, errors.New("nil variable")
	}

	msgs, err := s.cli.Recv()
	if err != nil {
		return nil, err
	}

	var dms []*driver.Message
	for _, msg := range msgs.Messages {
		dms = append(dms, &driver.Message{
			Body:     msg.Body,
			Metadata: msg.Header,
		})
	}
	return dms, nil
}

// SendAcks implements driver.Subscription.SendAcks.
func (s *subscription) SendAcks(ctx context.Context, ids []driver.AckID) error {
	// Ack is a no-op.
	return nil
}

// CanNack implements driver.CanNack.
func (s *subscription) CanNack() bool { return false }

// SendNacks implements driver.Subscription.SendNacks. It should never be called
// because we return false for CanNack.
func (s *subscription) SendNacks(ctx context.Context, ids []driver.AckID) error {
	panic("unreachable")
}

// IsRetryable implements driver.Subscription.IsRetryable.
func (s *subscription) IsRetryable(error) bool { return false }

// As implements driver.Subscription.As.
func (s *subscription) As(i interface{}) bool {
	return false
}

// ErrorAs implements driver.Subscription.ErrorAs
func (*subscription) ErrorAs(error, interface{}) bool {
	return false
}

// ErrorCode implements driver.Subscription.ErrorCode
func (*subscription) ErrorCode(err error) gcerrors.ErrorCode {
	if err == errNotExist {
		return gcerrors.NotFound
	}
	return gcerrors.Unknown
}

// Close implements driver.Subscription.Close.
func (*subscription) Close() error { return nil }
