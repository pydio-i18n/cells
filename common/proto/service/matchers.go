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

package service

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v5/common/errors"
)

// Matcher interface provides a way to filter idm objects with standard XXXSingleQueries.
type Matcher interface {
	// Matches tries to apply a *SingleQuery on an existing object
	Matches(object interface{}) bool
}

// MultiMatcher parses a Query and transform it to a recursive tree of Matches
type MultiMatcher struct {
	matchers  []Matcher
	Operation OperationType
}

// MatcherParser is a generic function to parse a protobuf into Matcher
type MatcherParser func(o *anypb.Any) (Matcher, error)

// Parse transforms input query into Matcher interfaces
func (mm *MultiMatcher) Parse(q *Query, parser MatcherParser) error {
	mm.Operation = q.Operation
	for _, an := range q.SubQueries {
		subQ := &Query{}
		if m, e := parser(an); e == nil {
			mm.matchers = append(mm.matchers, m)
		} else if e := anypb.UnmarshalTo(an, subQ, proto.UnmarshalOptions{}); e == nil {
			subM := &MultiMatcher{}
			if er := subM.Parse(subQ, parser); er != nil {
				return er
			}
			mm.matchers = append(mm.matchers, subM)
		} else {
			return errors.New("could not parse service.Query to MultiMatcher")
		}
	}
	return nil
}

// Matches implements the Matcher interface
func (mm *MultiMatcher) Matches(object interface{}) bool {
	var res []bool
	for _, m := range mm.matchers {
		res = append(res, m.Matches(object))
	}
	return ReduceQueryBooleans(res, mm.Operation)
}

// ReduceQueryBooleans combines multiple booleans depending on Operation Type
func ReduceQueryBooleans(results []bool, operation OperationType) bool {

	reduced := true
	if operation == OperationType_AND {
		// If one is false, it's false
		for _, b := range results {
			reduced = reduced && b
		}
	} else {
		// At least one must be true
		reduced = false
		for _, b := range results {
			reduced = reduced || b
			if b {
				break
			}
		}
	}
	return reduced
}
