package middleware

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"time"
)

// server log

func Logger(l logger.Logger) Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			fields := map[string]interface{}{
				"req":   fmt.Sprintf("%+v", req),
				"start": start,
			}
			l.Log(ctx, logger.InfoLevel, fields, "request log")
			rsp, err := handler(ctx, req)
			latency := time.Since(start).Seconds()
			fields = map[string]interface{}{
				"rsp":     fmt.Sprintf("%+v", rsp),
				"latency": latency,
			}
			l.Log(ctx, logger.InfoLevel, fields, "response log")
			if err != nil {
				fields = map[string]interface{}{
					"err":     fmt.Errorf("%+v", err),
					"latency": latency,
				}
				l.Log(ctx, logger.ErrorLevel, fields, "error log")
			}
			return rsp, err
		}
	}
}
