package broker

import (
	"fmt"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	pb "github.com/pydio/cells/v4/common/proto/broker"
)

type Handler struct {
	pb.UnimplementedBrokerServer
	broker broker.Broker
}

func NewHandler(b broker.Broker) *Handler {
	return &Handler{
		broker: b,
	}
}

func (h *Handler) Name() string {
	return common.ServiceGrpcNamespace_ + common.ServiceBroker
}

func (h *Handler) Publish(stream pb.Broker_PublishServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			fmt.Println("Error is ", err)
			return err
		}

		fmt.Println("Publishing ", req)

		if err := h.broker.Publish(stream.Context(), req.Topic, req.Messages); err != nil {
			return err
		}
	}
}

func (h *Handler) Subscribe(req *pb.SubscribeRequest, stream pb.Broker_SubscribeServer) error {
	// TODO v4 - manage unsubscription
	_, err := h.broker.Subscribe(stream.Context(), req.Topic, func(msg broker.Message) error {
		var target = &pb.Messages{}
		_, err := msg.Unmarshal(target)
		if err != nil {
			return err
		}
		if err := stream.Send(target); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	select {}

	return nil
}
