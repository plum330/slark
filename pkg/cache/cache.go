package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dtm-labs/rockscache"
	"github.com/go-slark/slark/pkg/sf"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type Cache struct {
	rocks  *rockscache.Client
	sf     *sf.SingleFlight
	err    error // not found error
	expiry time.Duration
}

type Option func(*Cache)

func Error(err error) Option {
	return func(c *Cache) {
		c.err = err
	}
}

func Expiry(expiry time.Duration) Option {
	return func(c *Cache) {
		c.expiry = expiry
	}
}

func New(redis *redis.Client, opts ...Option) *Cache {
	c := &Cache{
		rocks:  rockscache.NewClient(redis, rockscache.NewDefaultOptions()),
		sf:     sf.NewSingFlight(),
		err:    gorm.ErrRecordNotFound,
		expiry: time.Hour * 24 * 7,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Cache) Fetch(ctx context.Context, key string, v any, fn func(any) error) (bool, error) {
	var found bool
	data, err := c.sf.Do(key, func() (interface{}, error) {
		return c.rocks.Fetch2(ctx, key, c.expiry, func() (string, error) {
			err := fn(v)
			if err != nil {
				if errors.Is(err, c.err) {
					return "", nil
				}
				return "", err
			}
			found = true
			data, err := json.Marshal(v)
			return string(data), err
		})
	})
	if err != nil {
		return found, err
	}
	str, _ := data.(string)
	if len(str) == 0 {
		return found, nil
	}
	return found, json.Unmarshal([]byte(data.(string)), v)
}

/*
 key: db unique index key
 kf : db primary index key
 fn: query primary index by unique index
 f: query value by primary index
*/

func (c *Cache) FetchIndex(ctx context.Context, key string, kf func(any) string, v any, fn, f func(any) error) error {
	var pk any
	found, err := c.Fetch(ctx, key, &pk, fn)
	if err != nil {
		return err
	}
	if found {
		data, e := json.Marshal(v)
		if e != nil {
			return nil
		}
		_ = c.rocks.RawSet(ctx, kf(pk), string(data), c.expiry)
		return nil
	}
	_, err = c.Fetch(ctx, kf(pk), v, f)
	return err
}

func (c *Cache) Delete(key string) error {
	return c.rocks.TagAsDeleted(key)
}
