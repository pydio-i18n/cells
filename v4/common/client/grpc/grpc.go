package grpc

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	servercontext "github.com/pydio/cells/v4/common/server/context"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/pydio/cells/v4/common"
	clientcontext "github.com/pydio/cells/v4/common/client/context"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/service/context/ckeys"
	metadata2 "github.com/pydio/cells/v4/common/service/context/metadata"
)

var (
	mox = map[string]grpc.ClientConnInterface{}

	CallTimeoutDefault = 10 * time.Minute
	CallTimeoutShort   = 1 * time.Second
)

func GetClientConnFromCtx(ctx context.Context, serviceName string, opt ...Option) grpc.ClientConnInterface {
	if ctx == nil {
		return NewClientConn(serviceName, opt...)
	}
	conn := clientcontext.GetClientConn(ctx)
	reg := servercontext.GetRegistry(ctx)
	opt = append(opt, WithClientConn(conn))
	opt = append(opt, WithRegistry(reg))
	return NewClientConn(serviceName, opt...)
}

// NewClientConn returns a client attached to the defaults.
func NewClientConn(serviceName string, opt ...Option) grpc.ClientConnInterface {
	opts := new(Options)
	for _, o := range opt {
		o(opts)
	}

	if c, o := mox[strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_)]; o {
		return c
	}

	if opts.ClientConn == nil || opts.DialOptions != nil {
		if opts.Registry == nil {
			debug.PrintStack()

			reg, err := registry.OpenRegistry(context.Background(), viper.GetString("registry"))
			if err != nil {
				return nil
			}

			opts.Registry = reg
		}

		opts.DialOptions = append([]grpc.DialOption{grpc.WithInsecure(), grpc.WithResolvers(NewBuilder(opts.Registry))}, opts.DialOptions...)
		conn, err := grpc.Dial("cells:///", opts.DialOptions...)
		if err != nil {
			return nil
		}
		opts.ClientConn = conn
	}

	return &clientConn{
		callTimeout:         opts.CallTimeout,
		ClientConnInterface: opts.ClientConn,
		serviceName:         common.ServiceGrpcNamespace_ + strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_),
	}
}

type clientConn struct {
	grpc.ClientConnInterface
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
	var cancel context.CancelFunc
	if cc.callTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
		// Todo v4: can we simply defer cancel() for Invoke ?
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	er := cc.ClientConnInterface.Invoke(ctx, method, args, reply, opts...)
	if er != nil && cancel != nil {
		cancel()
	}
	return er
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
	var cancel context.CancelFunc
	if cc.callTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	s, e := cc.ClientConnInterface.NewStream(ctx, desc, method, opts...)
	if e != nil && cancel != nil {
		cancel()
	}
	return s, e
}

// RegisterMock registers a stubbed ClientConnInterface for a given service
func RegisterMock(serviceName string, mock grpc.ClientConnInterface) {
	mox[serviceName] = mock
}
