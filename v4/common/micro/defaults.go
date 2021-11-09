/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
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

// Package defaults initializes the defaults GRPC clients and servers used by services
package defaults

import (
	"google.golang.org/grpc"
	"time"

	"github.com/micro/micro/v3/service/api"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/network/transport"
	"github.com/micro/micro/v3/service/registry"
	"github.com/micro/micro/v3/service/server"

	grpcclient "github.com/micro/micro/v3/service/client/grpc"
	grpcserver "github.com/micro/micro/v3/service/server/grpc"
)

var (
	clientOpts     []func() client.Option
	serverOpts     []func() server.Option
	serverHTTPOpts []func() api.Option

	DefaultStartupRegistry registry.Registry
)

func InitServer(opt ...func() server.Option) {
	serverOpts = append(serverOpts, opt...)

	server.DefaultServer = NewServer()
}

func InitHTTPServer(opt ...func() api.Option) {
	serverHTTPOpts = append(serverHTTPOpts, opt...)
}

func InitClient(opt ...func() client.Option) {
	clientOpts = append(clientOpts, opt...)

	client.DefaultClient = newClient()
}

func newClient(new ...client.Option) client.Client {
	var opts []client.Option

	opts = append(opts, client.PoolSize(1))
	opts = append(opts, client.RequestTimeout(10*time.Minute))
	// opts = append(opts, client.Wrap(NetworkClientWrapper))
	// opts = append(opts, client.Wrap(servicecontext.SpanClientWrapper))

	for _, o := range clientOpts {
		opts = append(opts, o())
	}

	opts = append(opts, new...)

	return grpcclient.NewClient(
		opts...,
	)
}

// NewClient returns a client attached to the defaults
func NewClient(new ...client.Option) client.Client {
	if len(new) == 0 {
		return client.DefaultClient
	}

	return newClient(new...)
}

// NewClient returns a client attached to the defaults
func NewClientConn(serviceName string) *grpc.ClientConn {
	conn, err := grpc.Dial(":8001", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil
	}

	return conn

}

// NewServer returns a server attached to the defaults
func NewServer(new ...server.Option) server.Server {
	opts := new
	for _, o := range serverOpts {
		opts = append(opts, o())
	}

	return grpcserver.NewServer(
		opts...,
	)
}

// func NewHTTPServer(new ...api.Option) server.Server {
// 	opts := new
// 	for _, o := range serverHTTPOpts {
//		opts = append(opts, o())
//	}

//	s := httpserver.NewServer(
//		":8002",
//		opts...,
//	)
//
//	return s
//}

func Registry() registry.Registry {
	return registry.DefaultRegistry
}

func StartupRegistry() registry.Registry {
	if DefaultStartupRegistry == nil {
		return Registry()
	}

	return DefaultStartupRegistry
}

func Transport() transport.Transport {
	return server.DefaultServer.Options().Transport
}

func Broker() broker.Broker {
	return broker.DefaultBroker
}
