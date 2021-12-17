package file

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"

	"github.com/pydio/cells/v4/common/utils/configx"
	"github.com/pydio/cells/v4/common/utils/filex"
)

type file struct {
	v       configx.Values
	path    string
	watcher *fsnotify.Watcher

	updates []chan struct{}
}

func New(path string) configx.Entrypoint {
	data, err := filex.Read(path)
	if err != nil {
		return nil
	}

	v := configx.New(configx.WithJSON())
	if err := v.Set(data); err != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// TODO v4 this should not return nil
		return nil
	}

	if err := watcher.Add(path); err != nil {
		// TODO v4 this should not return nil
		return nil
	}

	return &file{
		v:       v,
		path:    path,
		watcher: watcher,
	}
}

func (f *file) watch() {
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				data, err := filex.Read(event.Name)
				if err != nil {
					continue
				}

				if err := f.v.Set(data); err != nil {
					continue
				}

				for _, ch := range f.updates {
					ch <- struct{}{}
				}
			}
		}
	}
}

func (f *file) Get() configx.Value {
	return f.v.Get()
}

func (f *file) Set(data interface{}) error {
	return filex.Save(f.path, data)
}

func (f *file) Val(path ...string) configx.Values {
	return &values{Values: f.v.Val(path...), path: f.path}
}

func (f *file) Del() error {
	return fmt.Errorf("not implemented")
}

func (f *file) Watch(path ...string) (configx.Receiver, error) {
	ch := make(chan struct{})

	f.updates = append(f.updates, ch)

	// For the moment do nothing
	return &receiver{
		ch: ch,
		p:  path,
		v:  f.v.Val(path...),
	}, nil
}

type receiver struct {
	ch chan struct{}
	p  []string
	v  configx.Values
}

func (r *receiver) Next() (configx.Values, error) {
	select {
	case <-r.ch:
		v := r.v.Val(r.p...)
		if bytes.Compare(v.Bytes(), r.v.Bytes()) != 0 {
			r.v = v
			return v, nil
		}
	}

	return nil, fmt.Errorf("could not retrieve data")
}

func (r *receiver) Stop() {
	close(r.ch)
}

type values struct {
	configx.Values
	path string
}

func (v *values) Set(data interface{}) error {
	if err := v.Values.Set(data); err != nil {
		return err
	}

	return filex.Save(v.path, v.Values.Val("#").Map())
}
