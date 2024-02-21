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
	sf    *sf.Group
	err   error // not found error
}

func New(redis *redis.Client, sf *sf.Group, err error) *Cache {
	return &Cache{
		rocks: rockscache.NewClient(redis, rockscache.NewDefaultOptions()),
		sf:    sf,
		err:   err,
	}
}

func (c *Cache) Fetch(ctx context.Context, key string, expire time.Duration, v any, fn func(any) error) error {
	_, err, _ := c.sf.Do(key, func() (interface{}, error) {
		return c.rocks.Fetch2(ctx, key, expire, func() (string, error) {
			err := fn(v)
			if err != nil {
				if errors.Is(err, c.err) {
					return "", nil
				}
				return "", err
			}
			data, _ := json.Marshal(v)
			return string(data), nil
		})
	})
	return err
}

func (c *Cache) Delete(key string) error {
	return c.rocks.TagAsDeleted(key)
}
