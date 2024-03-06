package logging

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/transport"
	"time"
)

const (
	ClientLog = iota + 1
	ServerLog
)

func Log(lt int, l logger.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				trans transport.Transporter
				ok    bool
			)
			if lt == ClientLog {
				trans, ok = transport.FromClientContext(ctx)
			} else if lt == ServerLog {
				trans, ok = transport.FromServerContext(ctx)
			}
			if !ok {
				return handler(ctx, req)
			}
			kind := trans.Kind()
			operation := trans.Operate()
			start := time.Now()
			fields := map[string]interface{}{
				"request":   fmt.Sprintf("%+v", req),
				"start":     start.Format(time.RFC3339),
				"operation": operation,
				"kind":      kind,
				"type":      lt,
			}
			l.Log(ctx, logger.DebugLevel, fields, "request log")
			rsp, err := handler(ctx, req)
			fields = map[string]interface{}{
				"latency":   time.Since(start).Milliseconds(),
				"end":       time.Now().Format(time.RFC3339),
				"operation": operation,
				"kind":      kind,
				"type":      lt,
			}
			var level uint
			if err != nil {
				fields["error"] = fmt.Errorf("%+v", err)
				level = logger.ErrorLevel
			} else {
				fields["response"] = fmt.Sprintf("%+v", rsp)
				level = logger.DebugLevel
			}
			l.Log(ctx, level, fields, "response log")
			return rsp, err
		}
	}
}
