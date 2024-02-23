package lock

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"time"
)

// redis lock

type Lock struct {
	redis      *redis.Client
	key, value string
	expire     int64
}

func New(key string, expire int64, redis *redis.Client) *Lock {
	return &Lock{
		redis:  redis,
		key:    key,
		value:  xid.New().String(),
		expire: expire,
	}
}

func (l *Lock) Lock(ctx context.Context) (bool, error) {
	//result, err := l.redis.SetNX(ctx, l.key, l.value, time.Duration(l.expire*1000+500) * time.Millisecond).Result()
	src := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
    		redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    		return "OK"
		else
    		return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
		end`
	result, err := redis.NewScript(src).Run(ctx, l.redis, []string{l.key}, []interface{}{l.value, time.Duration(l.expire*1000+500) * time.Millisecond}...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	reply, _ := result.(string)
	return reply == "OK", nil
}

func (l *Lock) Unlock(ctx context.Context) (bool, error) {
	src := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
    		return redis.call("DEL", KEYS[1])
		else
    		return 0
		end`
	result, err := redis.NewScript(src).Run(ctx, l.redis, []string{l.key}, l.value).Result()
	if err != nil {
		return false, err
	}
	reply, _ := result.(int64)
	return reply == 1, nil
}
