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

package grpc

import (
	"context"
	"github.com/pydio/cells/v4/common/utils/queue"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/pydio/cells/v4/common/proto/tree"
)

func TestEventsSubscriber_Handle(t *testing.T) {

	Convey("Test Events Subscriber", t, func() {

		out := make(chan *queue.TypeWithContext[*tree.NodeChangeEvent])
		subscriber := &EventsSubscriber{
			outputChannel: out,
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		var output *queue.TypeWithContext[*tree.NodeChangeEvent]
		go func() {
			defer wg.Done()
			for e := range out {
				if e != nil {
					output = e
				} else {
					return
				}
			}
		}()

		ctx := context.Background()
		ev := &tree.NodeChangeEvent{
			Type:   tree.NodeChangeEvent_CREATE,
			Source: &tree.Node{},
		}
		subscriber.Handle(ctx, ev)
		close(out)

		wg.Wait()

		So(output, ShouldResemble, &queue.TypeWithContext[*tree.NodeChangeEvent]{
			Ctx:      ctx,
			Original: ev,
		})

	})

}
