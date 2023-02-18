package recovery

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
)

func Recovery(l logger.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			defer func() {
				if e := recover(); e != nil {
					//buf := make([]byte, 64<<10) // buf size : 64k
					//n := runtime.Stack(buf, false)
					//buf = buf[:n]
					v, ok := e.(error)
					if ok && errors.HasStack(v) {
						err = v
					} else {
						err = errors.New(errors.PanicCode, errors.Panic, fmt.Sprintf("%+v", e))
					}
					fields := map[string]interface{}{
						"req": fmt.Sprintf("%+v", req),
						//"error": fmt.Sprintf("%s", buf),
						"error": fmt.Sprintf("%+v", err),
					}
					l.Log(ctx, logger.ErrorLevel, fields, "recover")
					err = errors.InternalServer(errors.Panic, errors.Panic)
				}
			}()
			rsp, err = handler(ctx, req)
			return rsp, err
		}
	}
}
