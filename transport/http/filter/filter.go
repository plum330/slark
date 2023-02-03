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

func HandleMiddlewares(mw ...middleware.Middleware) http.HandlerFunc {
	middle := middleware.HandleMiddleware(mw...)
	return func(w http.ResponseWriter, r *http.Request) {
		next := func(ctx context.Context, req interface{}) (interface{}, error) {
			var err error
			// TODO ....
			return w, err
		}
		_, _ = middle(next)(r.Context(), r)

	}
}
