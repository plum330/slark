package config

import (
	"github.com/go-slark/slark/config/source/file"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/pkg/routine"
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
	changes   []func(*Config) // 变动callback
	watchers  map[string][]func(*Config)
	format    string
	delimiter string
	src       Source
}

func New() *Config {
	c := &Config{
		changed:   make(map[string]any),
		l:         sync.RWMutex{},
		cached:    sync.Map{},
		delimiter: ".",
		format:    "toml",
		src:       file.NewFile(""), // TODO 默认文件路径
		changes:   make([]func(*Config), 0),
		watchers:  make(map[string][]func(*Config)),
	}
	return c
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
		for _, change := range c.changes {
			change(c)
		}
		c.l.RUnlock()
		for range c.src.Watch() {
			cfg, err = c.src.Load()
			if err != nil {
				continue
			}
			_ = c.load(cfg)
			c.l.RLock()
			for _, change := range c.changes {
				change(c)
			}
			c.l.RUnlock()
		}
	})
	return nil
}

func (c *Config) load(data []byte) error {
	cfg := make(map[string]any)
	err := encoding.GetCodec(c.format).Unmarshal(data, &cfg)
	if err != nil {
		return err
	}
	c.apply(cfg)
	return nil
}

func (c *Config) set(key, value string) {
	paths := strings.Split(key, c.delimiter)
	lastKey := paths[len(paths)-1]
	m := search(c.changed, paths[:len(paths)-1])
	m[lastKey] = value
	c.apply(m)
}

func (c *Config) apply(cfg map[string]any) {
	c.l.Lock()
	defer c.l.Unlock()
	changes := make(map[string]any)
	merge(c.changed, cfg)
	data := find(c.changed, "", c.delimiter)
	for k, v := range data {
		vv, ok := c.cached.Load(k)
		if ok && !reflect.DeepEqual(vv, v) {
			changes[k] = v
		}
		c.cached.Store(k, v)
	}
	if len(changes) > 0 {
		c.notify(changes)
	}
}

func (c *Config) notify(changes map[string]any) {
	var changedWatchPrefixMap = map[string]struct{}{}

	for watchPrefix := range c.watchers {
		for key := range changes {
			// 前缀匹配即可
			// todo 可能产生错误匹配
			if strings.HasPrefix(key, watchPrefix) {
				changedWatchPrefixMap[watchPrefix] = struct{}{}
			}
		}
	}

	for changedWatchPrefix := range changedWatchPrefixMap {
		for _, handle := range c.watchers[changedWatchPrefix] {
			go handle(c)
		}
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
	m := copier(c.changed, paths[:len(paths)-1]...)
	data = m[paths[len(paths)-1]]
	c.cached.Store(key, data)
	return data
}

func (c *Config) Get(key string) any {
	return c.find(key)
}

func (c *Config) GetString(key string) string {
	return cast.ToString(c.Get(key))
}
