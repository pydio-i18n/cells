package registry

import (
	"context"
	"fmt"
	"github.com/micro/micro/v3/service/registry"
)

type Registry interface{
	Register(Service) error
	Deregister(Service) error
	GetService(string) ([]Service, error)
	ListServices() ([]Service, error)
}

var defaultURLMux = &URLMux{}

// DefaultURLMux returns the URLMux used by OpenTopic and OpenSubscription.
//
// Driver packages can use this to register their TopicURLOpener and/or
// SubscriptionURLOpener on the mux.
func DefaultURLMux() *URLMux {
	return defaultURLMux
}

// OpenTopic opens the Topic identified by the URL given.
// See the URLOpener documentation in driver subpackages for
// details on supported URL formats, and https://gocloud.dev/concepts/urls
// for more information.
func OpenRegistry(ctx context.Context, urlstr string) (Registry, error) {
	return defaultURLMux.OpenRegistry(ctx, urlstr)
}

type reg struct{
	registry.Registry
}

func New(r registry.Registry) Registry {
	return &reg{
		Registry: r,
	}
}

func (r *reg) Register(s Service) error {
	var p *registry.Service
	if ok := s.As(&p); !ok {
		return fmt.Errorf("not a service")
	}
	return r.Registry.Register(p)
}

func (r *reg) Deregister(s Service) error {
	var p *registry.Service
	if ok := s.As(&p); !ok {
		return fmt.Errorf("not a service")
	}
	return r.Registry.Deregister(p)
}

func (r *reg) ListServices() ([]Service, error) {
	var services []Service

	ss, err := r.Registry.ListServices()
	if err != nil {
		return nil, err
	}
	
	for _, s := range ss {
		services = append(services, &service{
			s: s,
		})
	}
	return services, nil
}

func (r *reg) GetService(name string) ([]Service, error) {
	var services []Service

	ss, err := r.Registry.GetService(name)
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		services = append(services, &service{
			s: s,
		})
	}
	return services, nil
}