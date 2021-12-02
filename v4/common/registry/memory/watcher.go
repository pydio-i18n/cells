package memory

import (
	"errors"
	"github.com/pydio/cells/v4/common/registry"
)

type watcher struct {
	id   string
	wo   registry.WatchOptions
	res  chan registry.Result
	exit chan bool
}

func (m *watcher) Next() (registry.Result, error) {
	for {
		select {
		case r := <-m.res:
			if r.Service() == nil {
				continue
			}

			// TODO v4
			//if len(m.wo.Service) > 0 && m.wo.Service != r.Service.Name {
			//	continue
			//}

			return r, nil
		case <-m.exit:
			return nil, errors.New("watcher stopped")
		}
	}
}

func (m *watcher) Stop() {
	select {
	case <-m.exit:
		return
	default:
		close(m.exit)
	}
}

type result struct {
	action string
	service registry.Service
}

func (r *result) Action() string {
	return r.action
}

func (r *result) Service() registry.Service {
	return r.service
}
