package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

// db cache

type Cache struct {
	rocks  *rockscache.Client
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

func New(redis redis.UniversalClient, opts ...Option) *Cache {
	c := &Cache{
		rocks:  rockscache.NewClient(redis, rockscache.NewDefaultOptions()),
		err:    nil,
		expiry: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Cache) Fetch(ctx context.Context, key string, v any, fn func(any) error) error {
	data, err := c.rocks.Fetch2(ctx, key, c.expiry, func() (string, error) {
		err := fn(v)
		if err != nil {
			if errors.Is(err, c.err) {
				return "", nil
			}
			return "", err
		}
		data, err := json.Marshal(v)
		return string(data), err
	})
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return c.err
	}
	return json.Unmarshal([]byte(data), v)
}

/*
 v: query object
 key: db unique index key
 pkf : db primary index key
 query: query primary index by unique index
 pkQuery: query value by primary index
*/

func (c *Cache) FetchIndex(ctx context.Context, v any, key string, pkf func(any) string, query func(any) (any, error), pkQuery func(any, any) error) error {
	var found bool
	// query primary key
	pk, e := c.rocks.Fetch2(ctx, key, c.expiry, func() (string, error) {
		pk, err := query(v)
		if err != nil {
			if errors.Is(err, c.err) {
				return "", nil
			}
			return "", err
		}
		found = true
		data, _ := json.Marshal(v)
		_ = c.rocks.RawSet(ctx, pkf(pk), string(data), c.expiry)
		data, _ = json.Marshal(pk)
		return string(data), nil
	})
	if e != nil {
		return e
	}
	if found {
		return nil
	}
	return c.Fetch(ctx, pkf(pk), v, func(v any) error {
		return pkQuery(pk, v)
	})
}

// delete update

func (c *Cache) Exec(ctx context.Context, v any, f func(any) error, keys ...string) error {
	err := f(v)
	if err != nil {
		return err
	}
	return c.Deletes(ctx, keys)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.rocks.TagAsDeleted2(ctx, key)
}

func (c *Cache) Deletes(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rocks.TagAsDeletedBatch2(ctx, keys)
}
