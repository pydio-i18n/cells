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

package runtime

// emptyRuntime is used as default for runtime
type emptyRuntime struct{}

func (e *emptyRuntime) AllSettings() map[string]interface{} {
	return map[string]interface{}{}
}

func (e *emptyRuntime) GetBool(key string) bool {
	return false
}

func (e *emptyRuntime) GetString(key string) string {
	return ""
}

func (e *emptyRuntime) GetStringSlice(key string) []string {
	return []string{}
}

func (e *emptyRuntime) IsSet(key string) bool {
	return false
}

func (e *emptyRuntime) Set(key string, value interface{}) {

}

func (e *emptyRuntime) SetDefault(key string, value interface{}) {

}
