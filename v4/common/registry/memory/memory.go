package memory

import (
	"context"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pydio/cells/v4/common/registry"
)

var (
	scheme        = "memory"
	shared        registry.Registry
	sharedOnce    = &sync.Once{}
	sendEventTime = 10 * time.Millisecond
)

type URLOpener struct{}

func init() {
	o := &URLOpener{}
	registry.DefaultURLMux().Register(scheme, o)
}

func (o *URLOpener) OpenURL(ctx context.Context, u *url.URL) (registry.Registry, error) {
	if u.Query().Get("cache") == "shared" {
		sharedOnce.Do(func() {
			shared = newMemory()
		})

		return shared, nil
	}
	return newMemory(), nil
}

type memory struct {
	services []registry.Service
	nodes    []registry.Node

	sync.RWMutex
	watchers map[string]*watcher
}

func newMemory() registry.Registry {
	return &memory{
		watchers: make(map[string]*watcher),
	}
}

func (m *memory) StartService(name string) error {
	s, err := m.GetService(name)
	if err != nil {
		return err
	}

	return s.Start()
}

func (m *memory) StopService(name string) error {
	s, err := m.GetService(name)
	if err != nil {
		return err
	}

	return s.Stop()
}

func (m *memory) RegisterService(service registry.Service) error {
	for k, v := range m.services {
		if v.Name() == service.Name() {
			// TODO v4 merge nodes here
			m.services[k] = service
			go m.sendEvent(&result{action: "update", service: service})
			return nil
		}
	}

	// not found - adding it
	go m.sendEvent(&result{action: "create", service: service})
	m.services = append(m.services, service)

	return nil
}

func (m *memory) DeregisterService(service registry.Service) error {
	for k, v := range m.services {
		if service.Name() == v.Name() {
			m.services = append(m.services[:k], m.services[k+1:]...)
			go m.sendEvent(&result{action: "delete", service: service})
		}
	}
	return nil
}

func (m *memory) GetService(s string) (registry.Service, error) {
	for _, v := range m.services {
		if s == v.Name() {
			return v, nil
		}
	}
	return nil, os.ErrNotExist
}

func (m *memory) ListServices() ([]registry.Service, error) {
	return m.services, nil
}

func (m *memory) WatchServices(opts ...registry.WatchOption) (registry.Watcher, error) {
	// parse the options, fallback to the default domain
	var wo registry.WatchOptions
	for _, o := range opts {
		o(&wo)
	}

	// construct the watcher
	w := &watcher{
		exit: make(chan bool),
		res:  make(chan registry.Result),
		id:   uuid.New().String(),
		wo:   wo,
	}

	m.Lock()
	m.watchers[w.id] = w
	m.Unlock()

	return w, nil
}

func (m *memory) As(interface{}) bool {
	return false
}

func (m *memory) sendEvent(r registry.Result) {
	m.RLock()
	watchers := make([]*watcher, 0, len(m.watchers))
	for _, w := range m.watchers {
		watchers = append(watchers, w)
	}
	m.RUnlock()

	for _, w := range watchers {
		select {
		case <-w.exit:
			m.Lock()
			delete(m.watchers, w.id)
			m.Unlock()
		default:
			select {
			case w.res <- r:
			case <-time.After(sendEventTime):
			}
		}
	}
}
