package middleware

import (
	"context"
	"net/http"
)

type Handler func(ctx context.Context, req interface{}) (interface{}, error)

type Middleware func(Handler) Handler

func ComposeMiddleware(mws ...Middleware) Middleware {
	return func(handler Handler) Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			handler = mws[i](handler)
		}
		return handler
	}
}

type HTTPMiddleware func(handler http.Handler) http.Handler

func ComposeHTTPMiddleware(handler http.Handler, mws ...HTTPMiddleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func WrapMiddleware(mws ...Middleware) HTTPMiddleware {
	middle := ComposeMiddleware(mws...)
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
