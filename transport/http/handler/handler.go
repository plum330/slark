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
	err  error // propagate error for middleware
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

func (w *Wrapper) SetResponseWriter(rw http.ResponseWriter) {
	w.rw = rw
}

func (w *Wrapper) SetError(err error) {
	w.err = err
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
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				handler.ServeHTTP(w, r)
				wrapper, _ := w.(*Wrapper)
				var err error
				if wrapper.code > 0 {
					err = errors.New(wrapper.code, wrapper.err.Error(), wrapper.err.Error())
				}
				return wrapper.rw, err
			}
			_, _ = middle(next)(r.Context(), r)
		})
	}
}
