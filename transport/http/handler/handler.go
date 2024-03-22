package handler

import (
	"net/http"
)

// wrap http response writer

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
