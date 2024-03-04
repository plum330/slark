package ray

import (
	"context"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
)

type Config struct {
	Builder   func() string
	RequestID string
}

type Option func(*Config)

func WithBuilder(b func() string) Option {
	return func(cfg *Config) {
		cfg.Builder = b
	}
}

func WithRequestId(requestID string) Option {
	return func(cfg *Config) {
		cfg.RequestID = requestID
	}
}

func BuildRequestID(opts ...Option) middleware.Middleware {
	cfg := &Config{
		Builder: func() string {
			return utils.BuildRequestID()
		},
		RequestID: utils.TraceID,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			rid, _ := ctx.Value(cfg.RequestID).(string)
			if len(rid) == 0 {
				rid = cfg.Builder()
			}
			context.WithValue(ctx, cfg.RequestID, rid)
			return handler(ctx, req)
		}
	}
}
