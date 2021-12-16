package broker

import (
	"context"
	"fmt"
	"time"

	"testing"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestBroker(t *testing.T) {
	unsub, err := Subscribe(context.Background(), "test", func(msg Message) error {
		empty := &emptypb.Empty{}
		ctx, err := msg.Unmarshal(empty)
		fmt.Println("I've got a message ? ", ctx, empty, err)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer unsub()

	if err := Publish(context.Background(), "test", &emptypb.Empty{}); err != nil {
		fmt.Println("Publish error ", err)
		return
	}

	<-time.After(1 * time.Second)
}
