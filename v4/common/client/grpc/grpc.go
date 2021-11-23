package grpc

import (
	"context"
	"fmt"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"google.golang.org/grpc"
	"sync"
)

var (
	conn *grpc.ClientConn
	once = &sync.Once{}
)
// NewClient returns a client attached to the defaults
func NewClientConn(serviceName string) grpc.ClientConnInterface {
	var err error
	once.Do(func() {
		conn, err = grpc.Dial("cells://:8001/", grpc.WithInsecure())
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &clientConn{
		ClientConn: conn,
		serviceName: serviceName,
	}
}

type clientConn struct{
	*grpc.ClientConn
	serviceName string
}

// Invoke performs a unary RPC and returns after the response is received
// into reply.
func (cc *clientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	ctx = servicecontext.WithServiceName(ctx, cc.serviceName)
	return cc.ClientConn.Invoke(ctx, method, args, reply, opts...)
}

// NewStream begins a streaming RPC.
func (cc *clientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = servicecontext.WithServiceName(ctx, cc.serviceName)
	return cc.ClientConn.NewStream(ctx, desc, method, opts...)
}
