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

package service

import (
	"context"

	json "github.com/pydio/cells/v5/common/utils/jsonx"
)

type StorageDriver struct {
	Handler      any
	DefaultMetas []map[string]string
}

type StorageDrivers []StorageDriver

func (s *StorageDrivers) Register(f any, defaultMetas ...map[string]string) {
	*s = append(*s, StorageDriver{f, defaultMetas})
}

type StorageOptions struct {
	SupportedDrivers map[string]StorageDrivers
	Handler          any
	Migrator         MigratorFunc
	prefix           interface{}

	jsonMeta string
}

func (o *StorageOptions) Prefix(options *ServiceOptions) string {
	val := ""
	if o.prefix == nil {
		return val
	}
	switch v := o.prefix.(type) {
	case func(*ServiceOptions) string:
		val = v(options)
	case func(*StorageOptions) string:
		val = v(o)
	case string:
		val = v
	}
	return val
}

func (o *StorageOptions) ToMeta() string {
	if o.jsonMeta == "" {
		m := make(map[string]interface{})
		m["supportedDrivers"] = o.SupportedDrivers
		if o.Migrator != nil {
			m["hasMigrator"] = true
		}
		d, _ := json.Marshal(m)
		o.jsonMeta = string(d)
	}
	return o.jsonMeta
}

type MigratorStatus struct {
	Status string
	Total  int64
	Count  int64
}

type MigratorFunc func(topCtx, fromCtx, toCtx context.Context, dryRun bool, status chan MigratorStatus) (map[string]int, error)

type StorageOption func(options *StorageOptions)

// WithStoragePrefix sets a prefix to be used differently depending on driver name
func WithStoragePrefix(i interface{}) StorageOption {
	return func(options *StorageOptions) {
		options.prefix = i
	}
}

// WithStorageMigrator provides a Migrate function from one DAO to another
func WithStorageMigrator(d MigratorFunc) ServiceOption {
	return func(options *ServiceOptions) {
		options.StorageOptions.Migrator = d
	}
}

// WithNamedStorageDrivers allows supporting multiple set of drivers
func WithNamedStorageDrivers(name string, f StorageDrivers) ServiceOption {
	return func(o *ServiceOptions) {
		m := o.StorageOptions.SupportedDrivers
		if m == nil {
			m = make(map[string]StorageDrivers)
		}

		m[name] = f

		o.StorageOptions.SupportedDrivers = m
	}
}

// WithStorageDrivers adds a storage handler to the current service
func WithStorageDrivers(f StorageDrivers) ServiceOption {
	return WithNamedStorageDrivers("main", f)
}
