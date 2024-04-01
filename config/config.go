package config

import (
	"github.com/go-slark/slark/config/source/env"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/pkg/routine"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"golang.org/x/net/context"
	"reflect"
	"strings"
	"sync"
)

type Config struct {
	l         sync.RWMutex
	changed   map[string]any
	cached    sync.Map
	callback  []func()
	delimiter string
	src       Source
}

func New(opts ...Option) *Config {
	c := &Config{
		changed:   make(map[string]any),
		l:         sync.RWMutex{},
		cached:    sync.Map{},
		delimiter: ".",
		src:       env.New(),
		callback:  make([]func(), 0),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Config)

func Callback(callback []func()) Option {
	return func(c *Config) {
		c.callback = callback
	}
}

func WithSource(src Source) Option {
	return func(c *Config) {
		c.src = src
	}
}

func (c *Config) Load() error {
	cfg, err := c.src.Load()
	if err != nil {
		return err
	}
	err = c.load(cfg)
	if err != nil {
		return err
	}
	routine.GoSafe(context.TODO(), func() {
		c.l.RLock()
		for _, callback := range c.callback {
			callback()
		}
		c.l.RUnlock()
		for range c.src.Watch() {
			cfg, err = c.src.Load()
			if err != nil {
				continue
			}
			_ = c.load(cfg)
			c.l.RLock()
			for _, callback := range c.callback {
				callback()
			}
			c.l.RUnlock()
		}
	})
	return nil
}

func (c *Config) load(data []byte) error {
	cfg := make(map[string]any)
	err := encoding.GetCodec(c.src.Format()).Unmarshal(data, &cfg)
	if err != nil {
		return err
	}
	c.apply(cfg)
	return nil
}

func (c *Config) apply(cfg map[string]any) {
	c.l.Lock()
	defer c.l.Unlock()
	changes := make(map[string]any)
	merge(c.changed, cfg)
	data := spread(c.changed, "", c.delimiter)
	for k, v := range data {
		vv, ok := c.cached.Load(k)
		if ok && !reflect.DeepEqual(vv, v) {
			changes[k] = v
		}
		c.cached.Store(k, v)
	}
	if len(changes) > 0 {
		// TODO
	}
}

func (c *Config) find(key string) any {
	data, ok := c.cached.Load(key)
	if ok {
		return data
	}
	paths := strings.Split(key, c.delimiter)
	c.l.RLock()
	defer c.l.RUnlock()
	m := deepSearch(c.changed, paths[:len(paths)-1])
	data = m[paths[len(paths)-1]]
	c.cached.Store(key, data)
	return data
}

func (c *Config) Unmarshal(v any, key ...string) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     v,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	if len(key) == 0 {
		c.l.RLock()
		err = decoder.Decode(c.changed)
		c.l.RUnlock()
		return err
	}
	return decoder.Decode(c.find(key[0]))
}

func (c *Config) set(key, value string) {
	paths := strings.Split(key, c.delimiter)
	lastKey := paths[len(paths)-1]
	m := search(c.changed, paths[:len(paths)-1])
	m[lastKey] = value
	c.apply(m)
}

func (c *Config) Get(key string) any {
	return c.find(key)
}

func (c *Config) GetString(key string) string {
	return cast.ToString(c.Get(key))
}
