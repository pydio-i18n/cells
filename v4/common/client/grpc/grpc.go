package grpc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
)

var (
	conn *grpc.ClientConn
	once = &sync.Once{}
	mox  = map[string]grpc.ClientConnInterface{}
)

// NewClientConn returns a client attached to the defaults.
func NewClientConn(serviceName string) grpc.ClientConnInterface {

	if c, o := mox[strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_)]; o {
		return c
	}

	var err error
	once.Do(func() {
		conn, err = grpc.Dial("cells://:8001/", grpc.WithInsecure())
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &clientConn{
		ClientConn:  conn,
		serviceName: common.ServiceGrpcNamespace_ + strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_),
	}
}

type clientConn struct {
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

// RegisterMock registers a stubbed ClientConnInterface for a given service
func RegisterMock(serviceName string, mock grpc.ClientConnInterface) {
	mox[serviceName] = mock
}
