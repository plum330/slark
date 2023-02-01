package middleware

import (
	"context"
	"net/http"
)

type Handler func(ctx context.Context, req interface{}) (interface{}, error)

type Middleware func(Handler) Handler

func HandleMiddleware(handler Handler, mw ...Middleware) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}

type HandlerFunc func(handler http.HandlerFunc) http.HandlerFunc

func HandleFunc(handlerFunc http.HandlerFunc, mw ...HandlerFunc) http.HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		handlerFunc = mw[i](handlerFunc)
	}
	return handlerFunc
}
