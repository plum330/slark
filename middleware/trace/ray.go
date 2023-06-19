package trace

import (
	"context"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"net/http"
)

func BuildRequestID(opts ...utils.Option) middleware.HTTPMiddleware {
	cfg := &utils.Config{
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
