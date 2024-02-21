package handler

import (
	"context"
	utils "github.com/go-slark/slark/pkg"
	"net/http"
)

type Config struct {
	Builder   func() string
	RequestID string
}

type RIDOption func(*Config)

func WithBuilder(b func() string) RIDOption {
	return func(cfg *Config) {
		cfg.Builder = b
	}
}

func WithRequestId(requestID string) RIDOption {
	return func(cfg *Config) {
		cfg.RequestID = requestID
	}
}

func BuildRequestID(opts ...RIDOption) Middleware {
	cfg := &Config{
		Builder: func() string {
			return utils.BuildRequestID()
		},
		RequestID: utils.RayID,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(cfg.RequestID)
			if len(rid) == 0 {
				rid = cfg.Builder()
			}
			r.Header.Set(cfg.RequestID, rid)
			r = r.WithContext(context.WithValue(r.Context(), cfg.RequestID, rid))
			handler.ServeHTTP(w, r)
		})
	}
}
