package registry

import (
	"context"
)

type Registry interface{
	StartService(string) error
	StopService(string) error

	RegisterService(Service) error
	DeregisterService(Service) error
	GetService(string) (Service, error)
	ListServices() ([]Service, error)
	WatchServices(...WatchOption) (Watcher, error)
	As(interface{}) bool
}

type WatchOption func(WatchOptions) error

type WatchOptions interface{}

type Context interface {
	Context(context.Context)
}

type Watcher interface {
	Next() (Result, error)
	Stop()
}

type Result interface {
	Action() string
	Service() Service
}