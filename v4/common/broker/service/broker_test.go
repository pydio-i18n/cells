package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/pydio/cells/v4/common/proto/tree"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/server/stubs/discoverytest"
	"gocloud.dev/pubsub"
)

func TestServiceBroker(t *testing.T) {
	grpc.RegisterMock(common.ServiceBroker, discoverytest.NewBrokerService())

	subscription, _ := NewSubscription("test")

	wg := &sync.WaitGroup{}

	go func() {
		// defer wg.Done()

		defer subscription.Shutdown(context.Background())

		for {
			msg, err := subscription.Receive(context.Background())
			if err == io.EOF {
				return
			}

			ev := &tree.NodeChangeEvent{}
			if err := proto.Unmarshal(msg.Body, ev); err != nil {
				return
			}
			fmt.Println(ev)
			fmt.Println("The message received is ? ", string(msg.Body), err)

			wg.Done()
		}

	}()

	topic, _ := NewTopic("test")

	numMessages := 10000

	wg.Add(numMessages)

	msg := &tree.NodeChangeEvent{Source: &tree.Node{Path: "source"}, Target: &tree.Node{Path: "target"}}
	b, err := proto.Marshal(msg)
	if err != nil {
		fmt.Println("Error ", err)
		return
	}

	for i := 0; i < numMessages; i++ {
		go topic.Send(context.Background(), &pubsub.Message{
			Body: b,
		})
	}

	wg.Wait()

	topic.Shutdown(context.Background())
}
