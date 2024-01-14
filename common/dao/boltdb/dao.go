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

// Package boltdb provides tools for using Bolt as a standard persistence layer for services
package boltdb

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/pydio/cells/v4/common/dao"
	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/registry/util"
	"github.com/pydio/cells/v4/common/service/metrics"
	"github.com/pydio/cells/v4/common/utils/configx"
)

const Driver = "boltdb"

func init() {
	dao.RegisterDAODriver(Driver, NewDAO, func(ctx context.Context, driver, dsn string) dao.ConnDriver {
		return &boltdb{}
	})
}

// DAO defines the functions specific to the boltdb dao
type DAO interface {
	dao.DAO
	DB() *bolt.DB
	Compact(ctx context.Context, opts map[string]interface{}) (int64, int64, error)
}

// Handler for the main functions of the DAO
type Handler struct {
	dao.DAO
	runtimeCtx  context.Context
	statusInput chan map[string]interface{}
	metricsName string
}

// NewDAO creates a new handler for the boltdb dao
func NewDAO(ctx context.Context, driver string, dsn string, prefix string) (dao.DAO, error) {
	conn, err := dao.NewConn(ctx, driver, dsn)
	if err != nil {
		return nil, err
	}
	metricsName := ""
	if u, e := url.Parse(dsn); e == nil {
		metricsName = path.Base(path.Dir(u.Path))
	}

	return &Handler{
		DAO:         dao.AbstractDAO(conn, driver, dsn, prefix),
		runtimeCtx:  ctx,
		metricsName: metricsName,
	}, nil
}

// Init initialises the handler
func (h *Handler) Init(context.Context, configx.Values) error {
	return nil
}

// LocalAccess overrides DAO
func (h *Handler) LocalAccess() bool {
	return true
}

// DB returns the bolt DB object
func (h *Handler) DB() *bolt.DB {
	if h == nil {
		return nil
	}

	if conn, _ := h.GetConn(h.runtimeCtx); conn != nil {
		return conn.(*bolt.DB)
	}
	return nil
}

// As implements the registry.StatusReporter conversion
func (h *Handler) As(i interface{}) bool {
	if sw, ok := i.(*registry.StatusReporter); ok {
		*sw = h
		return true
	}
	return h.DAO.As(i)
}

// WatchStatus implements the StatusReport methods
func (h *Handler) WatchStatus() (registry.StatusWatcher, error) {
	if h.statusInput == nil {
		h.statusInput = make(chan map[string]interface{})
	}
	w := util.NewChanStatusWatcher(h, h.statusInput)
	c := time.NewTicker(time.Duration(10+rand.Intn(11)) * time.Second)
	go func() {
		h.sendStatus()
		for range c.C {
			h.sendStatus()
		}
	}()
	return w, nil
}

func (h *Handler) sendStatus() {
	if db := h.DB(); db != nil {
		if st, e := os.Stat(db.Path()); e == nil {
			metrics.GetMetrics().Tagged(map[string]string{"dsn": h.Name(), "service": h.metricsName}).Gauge("bolt_usage").Update(float64(st.Size()))
			h.statusInput <- map[string]interface{}{"Usage": st.Size()}
		}
	}
}

// Compact makes a copy of the current DB and replace it as a connection
func (h *Handler) Compact(ctx context.Context, opts map[string]interface{}) (old int64, new int64, err error) {
	db := h.DB()
	p := db.Path()
	if st, e := os.Stat(p); e == nil {
		old = st.Size()
	}
	dir, base := filepath.Split(p)
	ext := filepath.Ext(base)
	base = strings.TrimSuffix(base, ext)

	copyPath := filepath.Join(dir, base+"-compact-copy"+ext)
	copyDB, e := bolt.Open(copyPath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if e != nil {
		return 0, 0, e
	}
	if e := copyDB.Update(func(txW *bolt.Tx) error {
		return db.View(func(txR *bolt.Tx) error {
			return txR.ForEach(func(name []byte, b *bolt.Bucket) error {
				bW, e := txW.CreateBucketIfNotExists(name)
				if e != nil {
					return e
				}
				return copyValuesOrBucket(bW, b)
			})
		})
	}); e != nil {
		copyDB.Close()
		os.Remove(copyPath)
		return 0, 0, e
	}
	copyDB.Close()

	if e := h.CloseConn(ctx); e != nil {
		return 0, 0, e
	}
	bakPath := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, time.Now().Unix(), ext))
	if er := os.Rename(p, bakPath); er != nil {
		return 0, 0, er
	}
	if er := os.Rename(copyPath, p); er != nil {
		return 0, 0, er
	}
	if copyDB, e = bolt.Open(p, 0600, &bolt.Options{Timeout: 5 * time.Second}); e != nil {
		return 0, 0, e
	}
	h.SetConn(ctx, copyDB)
	if opts != nil {
		if clear, ok := opts["ClearBackup"]; ok {
			if c, o := clear.(bool); o && c {
				if er := os.Remove(bakPath); er != nil {
					err = er
				}
			}
		}
	}
	if st, e := os.Stat(p); e == nil {
		new = st.Size()
	}
	return
}

func copyValuesOrBucket(bW, bR *bolt.Bucket) error {
	return bR.ForEach(func(k, v []byte) error {
		if v == nil {
			newBW, e := bW.CreateBucketIfNotExists(k)
			if e != nil {
				return e
			}
			newBR := bR.Bucket(k)
			newBW.SetSequence(newBR.Sequence())
			return copyValuesOrBucket(newBW, newBR)
		} else {
			return bW.Put(k, v)
		}
	})
}
