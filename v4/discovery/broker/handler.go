package broker

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/micro/micro/v3/proto/broker"
	"github.com/micro/micro/v3/service/broker"
	"go.uber.org/zap"

	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/service/errors"
)

type Handler struct {
	sync.Mutex
	failed map[string][]*pb.Message
}

func (h *Handler) Publish(ctx context.Context, req pb.PublishRequest) error {
	//defer stream.Close()
	//for {
	//	req, err := stream.Recv()
	//	if err != nil {
	//		return err
	//	}

	// validate the request
	if req.Message == nil {
		return errors.BadRequest("broker.Broker.Publish", "Missing message")
	}

	// ensure the header is not nil
	if req.Message.Header == nil {
		req.Message.Header = map[string]string{}
	}

	// set any headers which aren't already set
	if md, ok := metadata.FromContext(ctx); ok {
		for k, v := range md {
			if _, ok := req.Message.Header[k]; !ok {
				req.Message.Header[k] = v
			}
		}
	}

	//log.Info("Received event from broker service")
	if err := broker.Publish(req.Topic, &broker.Message{
		Header: req.Message.Header,
		Body:   req.Message.Body,
	}); err != nil {
		return errors.InternalServerError("broker.Broker.Publish", err.Error())
	}

	//log.Info("Published event to the memory broker")
	//}

	return nil
}

func (h *Handler) Subscribe(ctx context.Context, req *pb.SubscribeRequest, stream pb.Broker_SubscribeStream) error {

	errChan := make(chan error, 1)
	var connId string
	if md, ok := metadata.FromContext(ctx); ok {
		if c, o := md["conn-id"]; o {
			connId = c
			qq := h.failedQueue(connId)
			if len(qq) > 0 {
				fmt.Println("[TMP LOG] Resending failed messages for conn-id", connId, len(qq))
				for _, m := range qq {
					if e := stream.Send(m); e != nil {
						return e
					}
				}
			}
		}
	}
	// message Broker to stream back messages from broker
	Broker := func(m *broker.Message) error {
		msg := &pb.Message{Header: m.Header, Body: m.Body}
		if err := stream.Send(msg); err != nil {
			h.queue(connId, msg)
			select {
			case errChan <- err:
				//fmt.Println("stream.Send got error and sent to errChan", err, req.Topic)
				return err
			default:
				//fmt.Println("stream.Send got error and return", err, req.Topic)
				return err
			}
		}
		return nil
	}

	log.Debug("Subscribing to topic", zap.String("topic", req.Topic))
	sub, err := broker.DefaultBroker.Subscribe(req.Topic, Broker, broker.Queue(req.Queue))
	if err != nil {
		return errors.InternalServerError("broker.Broker.Subscribe", err.Error())
	}
	defer func() {
		log.Debug("Unsubscribing from topic", zap.String("topic", req.Topic))
		sub.Unsubscribe()
	}()

	select {
	case <-ctx.Done():
		log.Debug("Context done for subscription to topic %s", zap.String("topic", req.Topic))
		return nil
	case err := <-errChan:
		log.Debug("Subscription error for topic", zap.String("topic", req.Topic), zap.Error(err))
		return err
	}
}

func (h *Handler) queue(connId string, message *pb.Message) {
	if connId == "" {
		return
	}
	h.Lock()
	defer h.Unlock()
	if h.failed == nil {
		h.failed = make(map[string][]*pb.Message)
	}
	h.failed[connId] = append(h.failed[connId], message)
}

func (h *Handler) failedQueue(connId string) (out []*pb.Message) {
	if connId == "" {
		return
	}
	h.Lock()
	defer h.Unlock()
	if h.failed == nil {
		return
	}
	if mm, o := h.failed[connId]; o {
		// Empty queue
		delete(h.failed, connId)
		return mm
	}
	return
}
