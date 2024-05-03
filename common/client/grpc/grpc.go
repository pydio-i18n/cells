/*
 * Copyright (c) 2019-2022. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package grpc

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client"
	clientcontext "github.com/pydio/cells/v4/common/client/context"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/runtime"
	"github.com/pydio/cells/v4/common/runtime/runtimecontext"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/ckeys"
	metadata2 "github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/service/metrics"

	_ "google.golang.org/grpc/xds"
)

type ctxBalancerFilterKey struct{}

var (
	CallTimeoutShort         = 1 * time.Second
	WarnMissingConnInContext = false
)

func DialOptionsForRegistry(reg registry.Registry, options ...grpc.DialOption) []grpc.DialOption {

	var clusterConfig *client.ClusterConfig
	config.Get("cluster").Default(&client.ClusterConfig{}).Scan(&clusterConfig)
	clientConfig := clusterConfig.GetClientConfig("grpc")

	backoffConfig := backoff.DefaultConfig

	return append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(NewBuilder(reg, clientConfig.LBOptions()...)),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 1 * time.Minute, Backoff: backoffConfig}),
		grpc.WithChainUnaryInterceptor(
			ErrorNoMatchedRouteRetryUnaryClientInterceptor(),
			ErrorFormatUnaryClientInterceptor(),
			servicecontext.SpanUnaryClientInterceptor(),
			MetaUnaryClientInterceptor(),
			otelgrpc.UnaryClientInterceptor(),
		),
		grpc.WithChainStreamInterceptor(
			ErrorNoMatchedRouteRetryStreamClientInterceptor(),
			ErrorFormatStreamClientInterceptor(),
			servicecontext.SpanStreamClientInterceptor(),
			MetaStreamClientInterceptor(),
			otelgrpc.StreamClientInterceptor(),
		),
		// grpc.WithDisableRetry(),
	}, options...)
}

func ResolveConn(ctx context.Context, serviceName string, opt ...Option) grpc.ClientConnInterface {
	if ctx == nil {
		return NewClientConn(serviceName, runtime.Cluster(), opt...)
	}
	conn := clientcontext.GetClientConn(ctx)
	if conn == nil && WarnMissingConnInContext {
		fmt.Println("Warning, ResolveConn could not find conn, will create a new one")
		debug.PrintStack()
	}
	var reg registry.Registry
	runtimecontext.Get(ctx, runtimecontext.RegistryKey, &reg)
	opt = append(opt, WithClientConn(conn))
	opt = append(opt, WithRegistry(reg))

	tenantName := "default"
	if mm, ok := metadata2.FromContextRead(ctx); ok {
		if p, o := mm[common.XPydioTenantUuid]; o {
			tenantName = p
		}
	}

	return NewClientConn(serviceName, tenantName, opt...)
}

// NewClientConn returns a client attached to the defaults.
func NewClientConn(serviceName string, tenantName string, opt ...Option) grpc.ClientConnInterface {
	opts := new(Options)
	for _, o := range opt {
		o(opts)
	}

	if c, o := mox[strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_)]; o {
		return c
	}

	if opts.ClientConn == nil || opts.DialOptions != nil {
		if opts.Registry == nil {
			reg, err := registry.OpenRegistry(context.Background(), runtime.RegistryURL())
			if err != nil {
				return nil
			}

			opts.Registry = reg
		}
		conn, err := grpc.Dial("xds://"+runtime.Cluster()+".cells.com/cells", DialOptionsForRegistry(opts.Registry, opts.DialOptions...)...)
		if err != nil {
			return nil
		}
		opts.ClientConn = conn
	}

	return &clientConn{
		callTimeout:         opts.CallTimeout,
		ClientConnInterface: opts.ClientConn,
		balancerFilter:      opts.BalancerFilter,
		serviceName:         common.ServiceGrpcNamespace_ + strings.TrimPrefix(serviceName, common.ServiceGrpcNamespace_),
		tenantName:          tenantName,
	}
}

type clientConn struct {
	grpc.ClientConnInterface
	serviceName    string
	tenantName     string
	callTimeout    time.Duration
	balancerFilter client.BalancerTargetFilter
}

// Invoke performs a unary RPC and returns after the response is received
// into reply.
func (cc *clientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	opts = append([]grpc.CallOption{
		grpc.WaitForReady(true),
	}, opts...)

	//if metadata.ValueFromIncomingContext(ctx, ckeys.TargetServiceName)
	ctx = metadata.AppendToOutgoingContext(ctx, ckeys.TargetServiceName, cc.serviceName)
	ctx = metadata.AppendToOutgoingContext(ctx, ckeys.TargetTenantName, cc.tenantName)

	var cancel context.CancelFunc
	if cc.callTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
	}
	if cc.balancerFilter != nil {
		ctx = context.WithValue(ctx, ctxBalancerFilterKey{}, cc.balancerFilter)
	}
	er := cc.ClientConnInterface.Invoke(ctx, method, args, reply, opts...)
	if er != nil && cancel != nil {
		cancel()
	}
	return er
}

var (
	clientRC  = map[string]float64{}
	clientRCL = sync.Mutex{}
)

// NewStream begins a streaming RPC.
func (cc *clientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	opts = append([]grpc.CallOption{
		grpc.WaitForReady(true),
	}, opts...)

	ctx = metadata.AppendToOutgoingContext(ctx, ckeys.TargetServiceName, cc.serviceName)
	ctx = metadata.AppendToOutgoingContext(ctx, ckeys.TargetTenantName, cc.tenantName)

	var cancel context.CancelFunc
	if cc.callTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cc.callTimeout)
	}
	if cc.balancerFilter != nil {
		ctx = context.WithValue(ctx, ctxBalancerFilterKey{}, cc.balancerFilter)
	}

	s, e := cc.ClientConnInterface.NewStream(ctx, desc, method, opts...)
	if e != nil && cancel != nil {
		cancel()
	}
	if e == nil {
		// Prepare gauges
		key := cc.serviceName + desc.StreamName
		scope := metrics.GetMetrics().Tagged(map[string]string{"target": cc.serviceName, "method": desc.StreamName})
		gauge := scope.Gauge("open_streams")
		pri := common.LogLevel == zapcore.DebugLevel
		if cc.serviceName == "pydio.grpc.broker" || cc.serviceName == "pydio.grpc.log" || cc.serviceName == "pydio.grpc.audit" ||
			cc.serviceName == "pydio.grpc.jobs" || cc.serviceName == "pydio.grpc.registry" ||
			desc.StreamName == "StreamChanges" || desc.StreamName == "PostNodeChanges" {
			pri = false
		}

		clientRCL.Lock()
		clientRC[key]++
		gauge.Update(clientRC[key])
		clientRCL.Unlock()
		ss := debug.Stack()
		go func() {
			select {
			case <-s.Context().Done():
				clientRCL.Lock()
				clientRC[key]--
				gauge.Update(clientRC[key])
				clientRCL.Unlock()
			case <-time.After(20 * time.Second):
				if pri {
					fmt.Println("==> Stream Not Closed After 20s", key)
					fmt.Print(string(ss))
				}
			}
		}()
	}
	return s, e
}
