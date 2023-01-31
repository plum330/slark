package middleware

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"runtime"
)

func Recovery() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var err error
			defer func() {
				if e := recover(); e != nil {
					buf := (&bytes.Buffer{}).Bytes()
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					fmt.Printf("%v: %+v\n%s\n", e, req, buf)
					err = errors.InternalServer("unknown error", "unknown error")
				}
			}()
			rsp, err := handler(ctx, req)
			return rsp, err
		}
	}
}
