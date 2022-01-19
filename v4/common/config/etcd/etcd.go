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

package etcd

import (
	"bytes"
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"

	configx "github.com/pydio/cells/v4/common/utils/configx"
)

type etcd struct {
	v configx.Values

	path string
	cli  *clientv3.Client

	updates []chan struct{}
}

func NewSource(ctx context.Context, cli *clientv3.Client, path string) configx.Entrypoint {
	v := configx.New(configx.WithJSON())

	m := &etcd{
		v:    v,
		cli:  cli,
		path: path,
	}

	go m.watch(ctx)

	return m
}

func (m *etcd) watch(ctx context.Context) {
	watcher := m.cli.Watch(ctx, m.path)

	for {
		select {
		case resp, ok := <-watcher:
			if !ok {
				return
			}
			for _, ev := range resp.Events {
				if err := m.v.Set(ev.Kv.Value); err != nil {
					continue
				}

				for _, ch := range m.updates {
					ch <- struct{}{}
				}
			}
		}
	}
}

func (m *etcd) Get() configx.Value {
	v := configx.New(configx.WithJSON())

	resp, _ := m.cli.Get(context.Background(), m.path, clientv3.WithLimit(1))

	for _, kv := range resp.Kvs {
		v.Set(kv.Value)
	}

	return v
}

func (m *etcd) Val(path ...string) configx.Values {
	return &values{Values: m.v.Val(path...), rootPath: m.path, cli: m.cli}
}

func (m *etcd) Set(data interface{}) error {
	resp, err := m.cli.Put(context.Background(), m.path, data.(string))
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func (m *etcd) Del() error {
	return fmt.Errorf("not implemented")
}

func (m *etcd) Watch(path ...string) (configx.Receiver, error) {
	ch := make(chan struct{})

	m.updates = append(m.updates, ch)

	// For the moment do nothing
	return &receiver{
		ch: ch,
		p:  path,
		v:  m.v.Val(path...),
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
	cli *clientv3.Client

	configx.Values
	rootPath string
}

func (v *values) Set(data interface{}) error {
	if err := v.Values.Set(data); err != nil {
		return err
	}

	_, err := v.cli.Put(context.Background(), v.rootPath, v.Values.Val("#").String())
	if err != nil {
		return err
	}

	return nil
}
