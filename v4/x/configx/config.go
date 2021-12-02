package configx

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	json "github.com/pydio/cells/v4/x/jsonx"

	"github.com/spf13/cast"
)

var (
	ErrNoMarshallerDefined = errors.New("no marshaller defined")
	ErrNoUnmarshalerDefined = errors.New("no unmarshaler defined")
)

type Scanner interface {
	Scan(interface{}) error
}

type Watcher interface {
	Watch(path ...string) (Receiver, error)
}

type Receiver interface {
	Next() (Values, error)
	Stop()
}

type Key interface{}

type Value interface {
	Default(interface{}) Value

	Bool() bool
	Bytes() []byte
	Int() int
	Int64() int64
	Duration() time.Duration
	String() string
	StringMap() map[string]string
	StringArray() []string
	Slice() []interface{}
	Map() map[string]interface{}

	Scanner
}

type KVStore interface {
	Get() Value
	Set(value interface{}) error
	Del() error
}

type Entrypoint interface {
	KVStore
	Val(path ...string) Values
}

type Values interface {
	Entrypoint
	Value
}

type Ref interface {
	Get() string
}

type Source interface {
	Entrypoint
	Watcher
}

func NewFrom(i interface{}) Values {
	c := New()
	c.Set(i)
	return c
}

// config is standard
type config struct {
	v    interface{}
	d    interface{} // Default
	r    *config     // Root
	k    []string    // Reference to key for re-assignment
	opts Options
}

func New(opts ...Option) Values {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	return &config{
		opts: options,
	}
}

func (v *config) get() interface{} {
	if v == nil {
		return nil
	}

	if v.v != nil {
		useDefault := false

		switch vv := v.v.(type) {
		case map[interface{}]interface{}:
			if ref, ok := vv["$ref"]; ok {
				vvv := v.r.Val(ref.(string)).Get()
				switch vvvv := vvv.(type) {
				case *config:
					return vvvv.get()
				default:
					return vvvv
				}
			}
		case map[string]interface{}:
			if ref, ok := vv["$ref"]; ok {
				vvv := v.r.Val(ref.(string)).Get()
				switch vvvv := vvv.(type) {
				case *config:
					return vvvv.get()
				default:
					return vvvv
				}
			}
		case string:
			if vv == "default" {
				useDefault = true
			}
		}

		if !useDefault {
			str, ok := v.v.(string)
			if ok {
				if d := v.opts.Decrypter; d != nil {
					b, err := d.Decrypt(str)
					if err != nil {
						return v.v
					}
					return string(b)
				}
			}
			return v.v
		}
	}

	if v.d != nil {
		switch vv := v.d.(type) {
		case map[string]interface{}:
			if ref, ok := vv["$ref"]; ok {
				vvv := v.r.Val(ref.(string)).Get()
				switch vvvv := vvv.(type) {
				case *config:
					return vvvv.get()
				default:
					return vvvv
				}
			}
		case *ref:
			vvv := v.r.Val(vv.Get()).Get()
			switch vvvv := vvv.(type) {
			case *config:
				return vvvv.get()
			default:
				return vvvv
			}
		}
		return v.d
	}

	return nil
}

// Get retrieve interface
func (v *config) Get() Value {
	if v.v == nil && v.d == nil {
		return nil
	}

	switch vv := v.v.(type) {
	case map[string]interface{}:
		if ref, ok := vv["$ref"]; ok {
			return v.r.Val(ref.(string)).Get()
		}
	case *ref:
		return v.r.Val(vv.Get()).Get()
	}

	return v
}

// Default value set
func (v *config) Default(i interface{}) Value {
	if v.d == nil {
		v.d = i
	}

	switch vv := v.v.(type) {
	case string:
		if vv == "default" {
			v.v = nil
		}
	}

	return v.Get()
}

// Set data in interface
func (v *config) Set(data interface{}) error {
	if v == nil {
		return fmt.Errorf("value doesn't exist")
	}

	if v.opts.Unmarshaler != nil {
		switch vv := data.(type) {
		case []byte:
			if len(vv) > 0 {
				if err := v.opts.Unmarshaler.Unmarshal(vv, &data); err != nil {
					return err
				}
			}
		}
	}

	switch d := data.(type) {
	case *config:
		data = d.v
	}

	if len(v.k) == 0 {
		v.v = data
		return nil
	}

	k := v.k[len(v.k)-1]
	pk := v.k[0 : len(v.k)-1]

	// Retrieve parent value
	p := v.r.Val(pk...)
	m := p.Map()
	if data == nil {
		delete(m, k)
	} else {
		if enc := v.opts.Encrypter; enc != nil {
			switch vv := data.(type) {
			case []byte:
				// Encrypting value
				str, err := enc.Encrypt(vv)
				if err != nil {
					return err
				}

				data = str
			case string:
				// Encrypting value
				str, err := enc.Encrypt([]byte(vv))
				if err != nil {
					return err
				}

				data = str
			}
		}

		m[k] = data
	}

	p.Set(m)

	v.v = data

	return nil
}

func (v *config) Del() error {
	if v == nil {
		return fmt.Errorf("value doesn't exist")
	}

	return v.Set(nil)
}

// Val values cannot retrieve lower values as it is final
func (v *config) Val(s ...string) Values {
	keys := StringToKeys(s...)

	// Need to do something for reference
	if len(keys) == 1 && keys[0] == "#" {
		if v.r != nil {
			return v.r
		}
		return v
	} else if len(keys) > 0 && keys[0] == "#" {
		keys = keys[1:]
	} else {
		keys = append(v.k, keys...)
	}

	root := v.r
	if v.r == nil {
		root = v
	}

	if len(keys) == 0 {
		return v
	}

	pk := keys

	// Looking for the specific key
	var current interface{} = root.Map()

	for _, pkk := range pk {
		switch cv := current.(type) {
		case map[interface{}]interface{}:
			c, ok := cv[pkk]
			if !ok {
				// The parent doesn't actually exist here, we return the nil value
				return &config{nil, nil, root, keys, v.opts}
			}

			current = c
		case map[string]interface{}:
			c, ok := cv[pkk]
			if !ok {
				// The parent doesn't actually exist here, we return the nil value
				return &config{nil, nil, root, keys, v.opts}
			}

			current = c
		case []interface{}:
			i, err := strconv.Atoi(pkk)
			if err != nil || i < 0 || i >= len(cv) {
				return &config{nil, nil, root, keys, v.opts}
			}

			c := cv[i]

			current = c
		default:
			return &config{nil, nil, root, keys, v.opts}
		}
	}

	return &config{current, nil, root, keys, v.opts}
}

// Scan to interface
func (c *config) Scan(val interface{}) error {
	v := c.get()
	if v == nil {
		return nil
	}

	marshaller := c.opts.Marshaller
	if marshaller == nil {
		return ErrNoMarshallerDefined
	}

	str, err := marshaller.Marshal(v)
	if err != nil {
		return err
	}

	unmarshaler := c.opts.Unmarshaler
	if unmarshaler == nil {
		return ErrNoUnmarshalerDefined
	}

	return unmarshaler.Unmarshal(str, val)
}

func (c *config) Bool() bool {
	v := c.get()
	if v == nil {
		return false
	}
	return cast.ToBool(v)
}
func (c *config) Bytes() []byte {
	v := c.get()
	if v == nil {
		return []byte{}
	}
	switch v := c.v.(type) {
	case []interface{}, map[string]interface{}:
		if m := c.opts.Marshaller; m != nil {
			data, err := m.Marshal(v)
			if err != nil {
				return []byte{}
			}

			return data
		}

		return []byte{}
	case string:
		// Need to handle it differently
		if v == "default" {
			c.v = nil
		}
	}
	return []byte(cast.ToString(v))
}
func (c *config) Int() int {
	v := c.get()
	if v == nil {
		return 0
	}
	return cast.ToInt(v)
}
func (c *config) Int64() int64 {
	v := c.get()
	if v == nil {
		return 0
	}
	return cast.ToInt64(v)
}
func (c *config) Duration() time.Duration {
	v := c.get()
	if v == nil {
		return 0 * time.Second
	}
	return cast.ToDuration(v)
}
func (c *config) String() string {
	v := c.get()

	switch v := c.v.(type) {
	case []interface{}, map[string]interface{}:
		if m := c.opts.Marshaller; m != nil {
			data, err := m.Marshal(v)
			if err != nil {
				return ""
			}

			return string(data)
		}

		return ""
	case string:
		// Need to handle it differently
		if v == "default" {
			c.v = nil
		}
	}

	return cast.ToString(v)
}
func (c *config) StringMap() map[string]string {
	v := c.get()
	if v == nil {
		return map[string]string{}
	}
	return cast.ToStringMapString(v)
}
func (c *config) StringArray() []string {
	v := c.get()
	if v == nil {
		return []string{}
	}
	return cast.ToStringSlice(c.get())
}
func (c *config) Slice() []interface{} {
	v := c.get()
	if v == nil {
		return []interface{}{}
	}
	return cast.ToSlice(c.get())
}
func (c *config) Map() map[string]interface{} {
	v := c.get()
	if v == nil {
		return map[string]interface{}{}
	}
	r, _ := cast.ToStringMapE(v)
	return r
}
func (c *config) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	c.v = m

	return nil
}

func (c *config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.v)
}
