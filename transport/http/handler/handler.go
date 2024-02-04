package handler

import (
	"net/http"
)

type Middleware func(handler http.Handler) http.Handler

func WrapMiddleware(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}
