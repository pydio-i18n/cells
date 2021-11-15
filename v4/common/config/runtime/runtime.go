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

package runtime

import (
	"github.com/spf13/viper"
	"regexp"
)

// IsLocal check if the environment runtime config is generated locally
func IsLocal() bool {
	return viper.GetString("config") == "local"
}

// IsRemote check if the environment runtime config is a remote server
func IsRemote() bool {
	return viper.GetString("config") == "remote" || viper.GetString("config") == "raft"
}

func IsRequired(serviceName string) bool {
	args := viper.GetStringSlice("args")
	if len(args) == 0 {
		return true
	}

	for _, arg := range viper.GetStringSlice("args"){
		re := regexp.MustCompile(arg)
		if re.MatchString(arg) {
			return true
		}
	}

	return false
}