package grpc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/service/context/ckeys"
	metadata2 "github.com/pydio/cells/v4/common/service/context/metadata"
)

var (
	conn *grpc.ClientConn
	once = &sync.Once{}
	mox  = map[string]grpc.ClientConnInterface{}

	CallTimeoutDefault = 10 * time.Minute
	CallTimeoutShort   = 1 * time.Second
)

// NewClientConn returns a client attached to the defaults.
func NewClientConn(serviceName string, opt ...Option) grpc.ClientConnInterface {
	opts := new(Options)
	opts.DialOptions = append(opts.DialOptions, grpc.WithInsecure())
	for _, o := range opt {
		o(opts)
	}

	if c, o := mox[strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_)]; o {
		return c
	}

	conn, err := grpc.Dial("cells:///", opts.DialOptions...)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &clientConn{
		callTimeout: opts.CallTimeout,
		ClientConn:  conn,
		serviceName: common.ServiceGrpcNamespace_ + strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_),
	}
}

type clientConn struct {
	*grpc.ClientConn
	serviceName string
	callTimeout time.Duration
}

// Invoke performs a unary RPC and returns after the response is received
// into reply.
func (cc *clientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	md := metadata.MD{}
	if lmd, ok := metadata2.FromContext(ctx); ok {
		for k, v := range lmd {
			if strings.HasPrefix(k, ":") {
				continue
			}
			md.Set(ckeys.CellsMetaPrefix+k, v)
		}
	}
	md.Set(ckeys.TargetServiceName, cc.serviceName)
	if cc.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
		defer cancel()
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return cc.ClientConn.Invoke(ctx, method, args, reply, opts...)
}

// NewStream begins a streaming RPC.
func (cc *clientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	md := metadata.MD{}
	if lmd, ok := metadata2.FromContext(ctx); ok {
		for k, v := range lmd {
			if strings.HasPrefix(k, ":") {
				continue
			}
			md.Set(ckeys.CellsMetaPrefix+k, v)
		}
	}
	md.Set(ckeys.TargetServiceName, cc.serviceName)
	if cc.callTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
		defer cancel()
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return cc.ClientConn.NewStream(ctx, desc, method, opts...)
}

// RegisterMock registers a stubbed ClientConnInterface for a given service
func RegisterMock(serviceName string, mock grpc.ClientConnInterface) {
	mox[serviceName] = mock
}
