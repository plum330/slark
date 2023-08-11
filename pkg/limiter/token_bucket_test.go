package limiter

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTBLimiter(t *testing.T) {
	options := &redis.Options{
		Network:     "tcp",
		Addr:        "192.168.3.13:2379",
		Password:    "CtHHQNbFkXpw33ew",
		DB:          5,
		IdleTimeout: time.Duration(10) * time.Second,
	}
	l := NewTBLimiter(5, 10, redis.NewClient(options), "limiter")
	var allow int
	for i := 0; i < 100; i++ {
		time.Sleep(time.Second / time.Duration(100))
		if l.AllowN(time.Now(), 5) {
			allow++
		}
	}
	assert.True(t, allow >= 15)
}
