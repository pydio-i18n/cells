/*
 * Copyright (c) 2019-2022. Abstrium SAS <team (at) pydio.com>
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

package util

import (
	"fmt"
	"time"

	pb "github.com/pydio/cells/v5/common/proto/registry"
	"github.com/pydio/cells/v5/common/registry"
	json "github.com/pydio/cells/v5/common/utils/jsonx"
)

func NewIntervalStatusWatcher(item registry.Item, interval time.Duration, callback func() (registry.Item, bool)) registry.StatusWatcher {

	s := &statusWatcher{
		item:     item,
		exit:     make(chan bool),
		ticker:   time.NewTicker(interval),
		callback: callback,
	}
	return s
}

type statusWatcher struct {
	item     registry.Item
	ticker   *time.Ticker
	exit     chan bool
	callback func() (registry.Item, bool)
}

func (s *statusWatcher) Next() (registry.Item, error) {
	for {
		select {
		case <-s.ticker.C:
			item, changed := s.callback()
			if !changed {
				continue
			}

			return item, nil
		case <-s.exit:
			return nil, fmt.Errorf("watcher stopped")
		}
	}
}

func (s *statusWatcher) Stop() {
	defer s.ticker.Stop()
	select {
	case <-s.exit:
		return
	default:
		close(s.exit)
	}
}

func NewIntervalStatsWatcher(item registry.Item, interval time.Duration, callback func() map[string]interface{}) registry.StatusWatcher {

	s := &statsWatcher{
		item:     item,
		exit:     make(chan bool),
		ticker:   time.NewTicker(interval),
		callback: callback,
	}
	return s
}

type statsWatcher struct {
	item     registry.Item
	ticker   *time.Ticker
	exit     chan bool
	callback func() map[string]interface{}
}

func (s *statsWatcher) Next() (registry.Item, error) {
	for {
		select {
		case <-s.ticker.C:
			ss := s.callback()
			js, _ := json.Marshal(ss)
			gen := &pb.Item{
				Id:       s.item.ID() + "-stats",
				Name:     "stats",
				Metadata: map[string]string{"Data": string(js)},
			}
			return ToGeneric(gen, &pb.Generic{Type: pb.ItemType_STATS}), nil
		case <-s.exit:
			return nil, fmt.Errorf("watcher stopped")
		}
	}
}

func (s *statsWatcher) Stop() {
	defer s.ticker.Stop()
	select {
	case <-s.exit:
		return
	default:
		close(s.exit)
	}
}
