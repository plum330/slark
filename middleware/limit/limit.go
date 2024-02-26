package limit

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/limiter"
)

type Limiter struct {
	limiter limiter.Limiter
}

type Option func(limiter *Limiter)

func WithLimiter(limiter limiter.Limiter) Option {
	return func(l *Limiter) {
		l.limiter = limiter
	}
}

func Limit(opts ...Option) middleware.Middleware {
	// TODO
	l := &Limiter{limiter: nil}
	for _, opt := range opts {
		opt(l)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			err := l.limiter.Pass()
			if err != nil {
				return nil, errors.New(430, "server rate limit", "SERVER_RATE_LIMIT")
			}
			rsp, err := handler(ctx, req)
			return rsp, err
		}
	}
}
