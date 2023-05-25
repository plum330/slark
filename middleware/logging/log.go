package logging

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"time"
)

// server log

func Log(l logger.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			fn := ctx.Value(utils.Method)
			fields := map[string]interface{}{
				"request": fmt.Sprintf("%+v", req),
				"start":   start.Format(time.RFC3339),
				"fn":      fn,
			}
			l.Log(ctx, logger.InfoLevel, fields, "request log")
			rsp, err := handler(ctx, req)
			fields = map[string]interface{}{
				"latency": time.Since(start).Seconds(),
				"fn":      fn,
			}
			var (
				level uint
				msg   string
			)
			if err != nil {
				fields["error"] = fmt.Errorf("%+v", err)
				level = logger.ErrorLevel
				msg = "error log"
			} else {
				fields["response"] = fmt.Sprintf("%+v", rsp)
				level = logger.InfoLevel
				msg = "response log"
			}
			l.Log(ctx, level, fields, msg)
			return rsp, err
		}
	}
}
