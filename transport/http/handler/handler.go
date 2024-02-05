package handler

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"net/http"
)

type Middleware func(handler http.Handler) http.Handler

func ComposeMiddleware(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func WrapMiddleware(mws ...middleware.Middleware) Middleware {
	middle := middleware.ComposeMiddleware(mws...)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				handler.ServeHTTP(w, r)
				return nil, nil
			}
			_, _ = middle(next)(r.Context(), r)
		})
	}
}
