package middleware

import (
	"context"
)

type Handler func(context.Context, interface{}) (interface{}, error)

type Middleware func(Handler) Handler

func ComposeMiddleware(mws ...Middleware) Middleware {
	return func(handler Handler) Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			handler = mws[i](handler)
		}
		return handler
	}
}
