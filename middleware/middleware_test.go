package middleware

import (
	"context"
	"fmt"
	"testing"
)

func TestMiddleware(t *testing.T) {
	_, _ = Handle(hello(), firstMiddleware(), secondMiddleware())(context.TODO(), "$$$")
}

// handler program
func hello() Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		fmt.Println("hello world")
		return nil, nil
	}
}

func firstMiddleware() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// do something in middleware before
			fmt.Println("1111111++++++++++")
			result, err := handler(ctx, req) // call next middleware / handler program
			// do something in middleware after
			fmt.Println("1111111---------")
			return result, err
		}
	}
}

func secondMiddleware() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			fmt.Println("2222222++++++++++")
			result, err := handler(ctx, req)
			fmt.Println("222222---------")
			return result, err
		}
	}
}
