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

package hooks

import "bytes"

var (
	pluginPaths []string
	templates   []TemplateFunc
	configPaths [][]string
)

// TemplateFunc is a function providing a stringer
type TemplateFunc func(site ...interface{}) (*bytes.Buffer, error)

// RegisterPluginTemplate adds a TemplateFunc to be called for each plugin
func RegisterPluginTemplate(fn TemplateFunc, watchConfigPath []string, pluginPaths ...string) error {
	pluginPaths = append(pluginPaths, pluginPaths...)
	templates = append(templates, fn)
	if len(watchConfigPath) > 0 {
		configPaths = append(configPaths, watchConfigPath)
	}
	return nil
}

func GetPluginPaths() []string {
	return pluginPaths
}

func GetTemplates() []TemplateFunc {
	return templates
}

func GetConfigPaths() [][]string {
	return configPaths
}
