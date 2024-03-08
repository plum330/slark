package handler

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"net/http"
)

type Wrapper struct {
	rw   http.ResponseWriter
	code int
}

func (w *Wrapper) WriteHeader(code int) {
	w.code = code
	w.rw.WriteHeader(http.StatusOK)
}

func (w *Wrapper) Header() http.Header {
	return w.rw.Header()
}

func (w *Wrapper) Write(data []byte) (int, error) {
	return w.rw.Write(data)
}

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
			wrapper, ok := w.(*Wrapper)
			if !ok {
				wrapper = &Wrapper{rw: w}
			}
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				handler.ServeHTTP(wrapper, r)
				var err error
				if wrapper.code != http.StatusOK {
					err = errors.New(wrapper.code, errors.UnknownReason, errors.UnknownReason)
				}
				return wrapper.rw, err
			}
			middle(next)(r.Context(), r)
		})
	}
}
