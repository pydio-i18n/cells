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
	"context"
	"time"

	"github.com/micro/micro/v3/service/broker"
	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	defaults "github.com/pydio/cells/v4/common/micro"
	pb "github.com/pydio/cells/v4/common/proto/broker"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

var (
	name    = "broker"
	address = ":8003"
)

type serviceBroker struct {
	Addrs   []string
	Client  pb.BrokerClient
	Stream  pb.Broker_PublishClient
	options broker.Options
}

func (b *serviceBroker) Address() string {
	return b.Addrs[0]
}

func (b *serviceBroker) Connect() error {
	stream, err := b.Client.Publish(context.TODO())
	if err != nil {
		return err
	}
	b.Stream = stream
	return nil
}

func (b *serviceBroker) Disconnect() error {
	return nil
}

func (b *serviceBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&b.options)
	}
	return nil
}

func (b *serviceBroker) Options() broker.Options {
	return b.options
}

func (b *serviceBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	err := b.Stream.SendMsg(&pb.PublishRequest{
		Topic: topic,
		Message: &pb.Message{
			Header: msg.Header,
			Body:   msg.Body,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *serviceBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	var options broker.SubscribeOptions
	for _, o := range opts {
		o(&options)
	}
	ctx := metadata.NewContext(context.Background(), map[string]string{"conn-id": uuid.New()})
	stream, err := b.Client.Subscribe(ctx, &pb.SubscribeRequest{
		Topic: topic,
		Queue: options.Queue,
	})
	if err != nil {
		return nil, err
	}

	sub := &serviceSub{
		topic:   topic,
		queue:   options.Queue,
		handler: handler,
		stream:  stream,
		closed:  make(chan bool),
		options: options,
	}

	go func() {
		for {
			select {
			case <-sub.closed:
				return
			default:
				if err := sub.run(); err != nil {
					stream, err := b.Client.Subscribe(ctx, &pb.SubscribeRequest{
						Topic: topic,
						Queue: options.Queue,
					})
					if err != nil {
						time.Sleep(time.Second)
						continue
					}
					// new stream
					sub.stream = stream
				}
			}
		}
	}()

	return sub, nil
}

func (b *serviceBroker) String() string {
	return "service"
}

func NewBroker(opts ...broker.Option) broker.Broker {
	var options broker.Options
	for _, o := range opts {
		o(&options)
	}

	addrs := options.Addrs
	if len(addrs) == 0 {
		addrs = []string{address}
	}

	// extract the client from the context, fallback to grpc
	var cli *grpc.ClientConn
	if c, ok := options.Context.Value(clientKey{}).(*grpc.ClientConn); ok {
		cli = c
	} else {
		cli = defaults.NewClientConn(common.ServiceBroker)
	}

	return &serviceBroker{
		Addrs:   addrs,
		Client:  pb.NewBrokerClient(cli),
		options: options,
	}
}
