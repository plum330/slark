package flow

import (
	"context"
	"errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/limit"
	"github.com/zeromicro/go-zero/core/load"
	"golang.org/x/time/rate"
	"time"
)

type Limiter interface {
	Pass() (func(err error), error)
}

// shedding

type Shedding struct {
	threshold int64
	shed      load.Shedder
}

type SheddingOption func(*Shedding)

func Threshold(threshold int64) SheddingOption {
	return func(s *Shedding) {
		s.threshold = threshold
	}
}

func NewShedding(opts ...SheddingOption) Limiter {
	sh := &Shedding{
		threshold: 900, // 0 - 1000 mill core
	}
	for _, opt := range opts {
		opt(sh)
	}
	sh.shed = load.NewAdaptiveShedder(load.WithCpuThreshold(sh.threshold))
	return sh
}

func (s *Shedding) Pass() (func(error), error) {
	promise, err := s.shed.Allow()
	if err != nil {
		return nil, err
	}
	f := func(err error) {
		if errors.Is(err, context.DeadlineExceeded) {
			promise.Fail()
		} else {
			promise.Pass()
		}
	}
	return f, nil
}

// max conn

type MaxConn struct {
	conn int
	pool *limit.Pool
}

type MaxConnOption func(*MaxConn)

func WitMaxConn(conn int) MaxConnOption {
	return func(c *MaxConn) {
		c.conn = conn
	}
}

func NewMaxConn(opts ...MaxConnOption) Limiter {
	mc := &MaxConn{conn: 1000}
	for _, opt := range opts {
		opt(mc)
	}
	mc.pool = limit.NewPool(mc.conn)
	return mc
}

func (c *MaxConn) Pass() (func(error), error) {
	allow := c.pool.Use()
	if !allow {
		return nil, errors.New("max conn overload")
	}
	err := c.pool.Back()
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err})
	}
	return func(err error) {}, nil
}

// rate limit

type RateLimit struct {
	limiter *rate.Limiter
}

type RateLimitOption func(*RateLimit)

func WithRateLimiter(limiter *rate.Limiter) RateLimitOption {
	return func(limit *RateLimit) {
		limit.limiter = limiter
	}
}

func NewRateLimiter(opts ...RateLimitOption) Limiter {
	r := &RateLimit{
		limiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 1000),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *RateLimit) Pass() (func(error), error) {
	if !r.limiter.AllowN(time.Now(), 1) {
		return nil, errors.New("rate limit")
	}
	return func(error) {}, nil
}
