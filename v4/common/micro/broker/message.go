package broker

import (
	"context"
	"github.com/micro/micro/v3/service/context/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/micro/micro/v3/service/broker"
)

type UnSubscriber func() error

type SubscriberHandler func(Message) error

type Message interface {
	Unmarshal(target proto.Message) (context.Context, error)
}

type SubMessage struct {
	*broker.Message
}

func (m *SubMessage) Unmarshal(target proto.Message) (context.Context, error) {
	if e := proto.Unmarshal(m.Body, target); e != nil {
		return nil, e
	}
	ctx := context.Background()
	if m.Header != nil {
		ctx = metadata.NewContext(ctx, m.Header)
	}
	return ctx, nil
}
