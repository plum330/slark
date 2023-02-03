package middleware

import (
	"context"
)

type Handler func(ctx context.Context, req interface{}) (interface{}, error)

type Middleware func(Handler) Handler

func HandleMiddleware(mw ...Middleware) Middleware {
	return func(handler Handler) Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			handler = mw[i](handler)
		}
		return handler
	}
}
