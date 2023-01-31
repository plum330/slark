package middleware

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"runtime"
)

func Recovery(l logger.Logger) Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var err error
			defer func() {
				if e := recover(); e != nil {
					buf := make([]byte, 64<<10) // buf:64k
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					fields := map[string]interface{}{
						"error": e,
						"req":   fmt.Sprintf("%+v", req),
						"stack": fmt.Sprintf("%s", buf),
					}
					l.Log(ctx, logger.ErrorLevel, fields, "recovery")
					err = errors.InternalServer("unknown error", "unknown error")
				}
			}()
			rsp, err := handler(ctx, req)
			return rsp, err
		}
	}
}
