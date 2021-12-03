package middleware

import (
	"github.com/pydio/cells/v4/common/registry"
)

type mockRegistry struct {
	r registry.Registry
}

func NewMockRegistry(r registry.Registry) registry.Registry {
	return &mockRegistry{
		r: r,
	}
}

func (r *mockRegistry) StartService(s string) error {
	return r.r.StartService(s)
}

func (r *mockRegistry) StopService(s string) error {
	return r.r.StopService(s)
}

func (r *mockRegistry) RegisterService(service registry.Service) error {
	return r.r.RegisterService(service)
}

func (r *mockRegistry) DeregisterService(service registry.Service) error {
	return r.r.DeregisterService(service)
}

func (r *mockRegistry) GetService(s string) (registry.Service, error) {
	return r.r.GetService(s)
}

func (r *mockRegistry) ListServices() ([]registry.Service, error) {
	return r.r.ListServices()
}

func (r *mockRegistry) WatchServices(option ...registry.WatchOption) (registry.Watcher, error) {
	return r.r.WatchServices(option...)
}

func (r *mockRegistry) As(i interface{}) bool {
	return r.r.As(i)
}