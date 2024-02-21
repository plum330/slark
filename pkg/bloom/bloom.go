package bloom

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/spaolacci/murmur3"
)

/*
 1. 位数组 redis bitmap
 2. 散列函数 MurmurHash3/CityHash
*/

type Bloom struct {
	bits  uint
	times uint
	key   string
	redis *redis.Client
}

type Option func(*Bloom)

// Times 散列次数
func Times(times uint) Option {
	return func(bloom *Bloom) {
		bloom.times = times
	}
}

// Bits 数组长度
func Bits(bits uint) Option {
	return func(bloom *Bloom) {
		bloom.bits = bits
	}
}

func New(key string, opts ...Option) *Bloom {
	bloom := &Bloom{
		key:   key,
		bits:  32,
		times: 14,
	}
	for _, opt := range opts {
		opt(bloom)
	}
	return bloom
}

func (b *Bloom) hash(data []byte) []uint {
	h := make([]uint, b.times)
	var i uint
	for ; i < b.times; i++ {
		h[i] = uint(murmur3.Sum64(append(data, byte(i)))) % b.bits
	}
	return h
}

func (b *Bloom) check(positions []uint) ([]uint, error) {
	args := make([]uint, 0, len(positions))
	for _, position := range positions {
		if position >= b.bits {
			return nil, errors.New("hash position out of range")
		}
		args = append(args, position)
	}
	return args, nil
}

func (b *Bloom) Set(ctx context.Context, data []byte) error {
	src := `
		for _, offset in ipairs(ARGV) do
			redis.call("setbit", KEYS[1], offset, 1)
		end
	`
	h := b.hash(data)
	return redis.NewScript(src).Run(ctx, b.redis, []string{b.key}, h).Err()
}

func (b *Bloom) Exist(ctx context.Context, data []byte) (bool, error) {
	src := `
		for _, offset in ipairs(ARGV) do
			if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
				return false
			end
		end
		return true
	`
	h := b.hash(data)
	result, err := redis.NewScript(src).Run(ctx, b.redis, []string{b.key}, h).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	exists, _ := result.(int64)
	return exists == 1, nil
}
