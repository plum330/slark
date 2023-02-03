package filter

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"net/http"
)

type Handler func(handler http.Handler) http.Handler

func Handle(handler http.Handler, mw ...Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}

func HandleFilters(mw ...middleware.Middleware) Handler {
	middle := middleware.HandleMiddleware(mw...)
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
