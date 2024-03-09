package limit

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/limit"
)

type Limiter struct {
	limiter limit.Limiter
}

type Option func(limiter *Limiter)

func WithLimiter(limiter limit.Limiter) Option {
	return func(l *Limiter) {
		l.limiter = limiter
	}
}

func Limit(opts ...Option) middleware.Middleware {
	l := &Limiter{limiter: limit.NewMaxConn()}
	for _, opt := range opts {
		opt(l)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			err := l.limiter.Pass()
			if err != nil {
				return nil, errors.ServerRateLimit("server rate limit", err.Error())
			}
			return handler(ctx, req)
		}
	}
}
