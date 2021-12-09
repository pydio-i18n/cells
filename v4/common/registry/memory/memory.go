package memory

import (
	"context"
	pb "github.com/pydio/cells/v4/common/proto/registry"
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
			shared = registry.NewRegistry(newMemory())
		})

		return shared, nil
	}
	return registry.NewRegistry(newMemory()), nil
}

type memory struct {
	register []registry.Item

	sync.RWMutex
	watchers map[string]*watcher
}

type options struct {
	itemType int
}

func newMemory() registry.Registry {
	return &memory{
		watchers: make(map[string]*watcher),
	}
}

func (m *memory) Start(item registry.Item) error {
	go m.sendEvent(&result{action: "start_request", item: item})
	return nil
}

func (m *memory) Stop(item registry.Item) error {
	go m.sendEvent(&result{action: "stop_request", item: item})

	return nil
}

func (m *memory) Register(item registry.Item) error {
	// Then register all services
	for k, v := range m.register {
		if v.Name() == item.Name() {
			m.register[k] = item
			go m.sendEvent(&result{action: "update", item: item})
			return nil
		}
	}

	// not found - adding it
	go m.sendEvent(&result{action: "create", item: item})
	m.register = append(m.register, item)

	return nil
}

func (m *memory) Deregister(item registry.Item) error {
	for k, v := range m.register {
		if item.Name() == v.Name() {
			m.register = append(m.register[:k], m.register[k+1:]...)
			go m.sendEvent(&result{action: "delete", item: item})
		}
	}
	return nil
}

func (m *memory) Get(s string, opts ...registry.Option) (registry.Item, error) {
	o := registry.Options{}
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	for _, v := range m.register {
		if s == v.Name() {
			return v, nil
		}
	}
	return nil, os.ErrNotExist
}

func (m *memory) List(opts ...registry.Option) ([]registry.Item, error) {
	var items []registry.Item

	o := registry.Options{}
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	if o.Type == pb.ItemType_ALL {
		return m.register, nil
	}

	for _, item := range m.register {
		switch o.Type {
		case pb.ItemType_SERVICE:
			if service, ok := item.(registry.Service); ok {
				items = append(items, service)
			}
		case pb.ItemType_NODE:
			if node, ok := item.(registry.Node); ok {
				items = append(items, node)
			}
		}
	}

	return items, nil
}

func (m *memory) Watch(opts ...registry.Option) (registry.Watcher, error) {
	// parse the options, fallback to the default domain
	var wo registry.Options
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