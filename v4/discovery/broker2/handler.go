package broker

import (
	"context"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

type Handler struct {
	pb.UnimplementedPublisherServer
	pb.UnimplementedSubscriberServer

	broker        broker.Broker
	subscriptions map[string][]*Message
}

func NewHandler(b broker.Broker) *Handler {
	return &Handler{
		broker: b,
	}
}

func (h *Handler) Name() string {
	return common.ServiceGrpcNamespace_ + common.ServiceBroker
}

func (h *Handler) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	var msgIds []string
	for _, msg := range req.GetMessages() {
		if err := h.broker.Publish(ctx, req.GetTopic(), msg); err != nil {
			continue
		}
		msgIds = append(msgIds, msg.GetMessageId())
	}
	return &pb.PublishResponse{MessageIds: msgIds}, nil
}

func (h *Handler) Pull(ctx context.Context, req *pb.PullRequest) (*pb.PullResponse, error) {
	sub, ok := h.subscriptions[req.GetSubscription()]
	if !ok {
		// TODO v4 - probably handler context instead
		h.broker.Subscribe(context.Background(), req.)
	}
	req.GetSubscription()
	return nil, status.Errorf(codes.Unimplemented, "method Pull not implemented")
}
