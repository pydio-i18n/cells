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

package service

import (
	"context"
	"net/url"
	"time"

	"github.com/pydio/cells/v4/common/registry"

	mregistry "github.com/micro/micro/v3/service/registry"
	pb "github.com/pydio/cells/v4/common/proto/registry"
	"google.golang.org/grpc"
)

var scheme = "grpc"

type URLOpener struct {
	*grpc.ClientConn
}

func init() {
	o := &URLOpener{}
	registry.DefaultURLMux().Register(scheme, o)
}

func (o *URLOpener) OpenURL(ctx context.Context, u *url.URL) (registry.Registry, error) {
	conn, err := grpc.Dial(u.Hostname()+":"+u.Port(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return NewRegistry(
		WithConn(conn),
	), nil
}

var (
	// The default service name
	DefaultService = "go.micro.registry"
)

type serviceRegistry struct {
	opts mregistry.Options
	// name of the registry
	name string
	// address
	address []string
	// client to call registry
	client pb.RegistryClient
}

func (s *serviceRegistry) callOpts() []grpc.CallOption {
	var opts []grpc.CallOption

	// set registry address
	//if len(s.address) > 0 {
	//	opts = append(opts, client.WithAddress(s.address...))
	//}

	// set timeout
	if s.opts.Timeout > time.Duration(0) {
		// opts = append(opts, grpc.client.WithRequestTimeout(s.opts.Timeout))
	}

	// add retries
	// TODO : charles' GUTS feeling :-)
	// opts = append(opts, client.WithRetries(10))

	return opts
}

func (s *serviceRegistry) Init(opts ...mregistry.Option) error {
	for _, o := range opts {
		o(&s.opts)
	}

	if len(s.opts.Addrs) > 0 {
		s.address = s.opts.Addrs
	}

	// extract the client from the context, fallback to grpc
	var conn *grpc.ClientConn
	if c, ok := s.opts.Context.Value(connKey{}).(*grpc.ClientConn); ok {
		conn = c
	} else {
		conn, _ = grpc.Dial(":8000")
	}

	s.client = pb.NewRegistryClient(conn)

	return nil
}

func (s *serviceRegistry) Options() mregistry.Options {
	return s.opts
}

func (s *serviceRegistry) StartService(name string) error {
	_, err := s.client.StartService(s.opts.Context, &pb.StartServiceRequest{
		Service: name,
	}, s.callOpts()...)
	if err != nil {
		return err
	}

	return nil
}

func (s *serviceRegistry) StopService(name string) error {
	_, err := s.client.StopService(s.opts.Context, &pb.StopServiceRequest{
		Service: name,
	}, s.callOpts()...)
	if err != nil {
		return err
	}

	return nil
}

func (s *serviceRegistry) RegisterService(srv registry.Service) error {
	_, err := s.client.RegisterService(s.opts.Context, ToProtoService(srv), s.callOpts()...)
	if err != nil {
		return err
	}

	return nil
}

func (s *serviceRegistry) DeregisterService(srv registry.Service) error {
	_, err := s.client.DeregisterService(s.opts.Context, ToProtoService(srv), s.callOpts()...)
	if err != nil {
		return err
	}
	return nil
}

func (s *serviceRegistry) GetService(name string) (registry.Service, error) {
	rsp, err := s.client.GetService(s.opts.Context, &pb.GetServiceRequest{
		Service: name,
	}, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	return ToService(rsp.Service), nil
}

func (s *serviceRegistry) ListServices() ([]registry.Service, error) {
	rsp, err := s.client.ListServices(s.opts.Context, &pb.ListServicesRequest{}, s.callOpts()...)
	if err != nil {
		return nil, err
	}

	services := make([]registry.Service, 0, len(rsp.Services))
	for _, service := range rsp.Services {
		services = append(services, ToService(service))
	}

	return services, nil
}

func (s *serviceRegistry) WatchServices(opts ...registry.WatchOption) (registry.Watcher, error) {
	var options mregistry.WatchOptions
	for _, o := range opts {
		o(&options)
	}
	if options.Context == nil {
		options.Context = context.TODO()
	}

	stream, err := s.client.WatchServices(options.Context, &pb.WatchServicesRequest{
		Service: options.Service,
	}, s.callOpts()...)

	if err != nil {
		return nil, err
	}

	return newWatcher(stream), nil
}

func (s *serviceRegistry) As(interface{}) bool {
	return false
}

func (s *serviceRegistry) String() string {
	return "service"
}

// NewRegistry returns a new registry service client
func NewRegistry(opts ...mregistry.Option) registry.Registry {
	var options mregistry.Options
	for _, o := range opts {
		o(&options)
	}

	var ctx context.Context
	var cancel context.CancelFunc

	ctx = options.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	ctx, cancel = context.WithCancel(ctx)

	options.Context = ctx

	// extract the client from the context, fallback to grpc
	var conn *grpc.ClientConn
	conn, ok := options.Context.Value(connKey{}).(*grpc.ClientConn)
	if !ok {
		conn, _ = grpc.Dial(":8000")
	}

	// service name. TODO: accept option
	name := DefaultService

	r := &serviceRegistry{
		opts:   options,
		name:   name,
		client: pb.NewRegistryClient(conn),
	}

	go func() {
		// Check the stream has a connection to the registry
		watcher, err := r.WatchServices()
		if err != nil {
			cancel()
			return
		}

		for {
			_, err := watcher.Next()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	return r
}
