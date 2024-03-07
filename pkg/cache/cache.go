package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dtm-labs/rockscache"
	"github.com/go-slark/slark/pkg/sf"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	rocks *rockscache.Client
	sf    *sf.SingleFlight
	err   error // not found error
}

func New(redis *redis.Client, sf *sf.SingleFlight, err error) *Cache {
	return &Cache{
		rocks: rockscache.NewClient(redis, rockscache.NewDefaultOptions()),
		sf:    sf,
		err:   err,
	}
}

func (c *Cache) Fetch(ctx context.Context, key string, expire time.Duration, v any, fn func(any) error) error {
	data, err := c.sf.Do(key, func() (interface{}, error) {
		return c.rocks.Fetch2(ctx, key, expire, func() (string, error) {
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
	})
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data.(string)), v)
}

/*
 key: db unique index key
 kf : db primary index key
 fn: query primary index by unique index
 f: query value by primary index
*/

func (c *Cache) FetchIndex(ctx context.Context, key string, expire time.Duration, kf func(any) string, v any, fn, f func(any) error) error {
	var pk any
	err := c.Fetch(ctx, key, expire, &pk, fn)
	if err != nil {
		return err
	}
	return c.Fetch(ctx, kf(pk), expire, v, f)
}

func (c *Cache) Delete(key string) error {
	return c.rocks.TagAsDeleted(key)
}
