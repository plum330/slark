package stat

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/zeromicro/go-zero/core/stat"
	"time"
)

func Stat(name string) middleware.Middleware {
	stat.DisableLog()
	metrics := stat.NewMetrics(name)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			defer func() {
				metrics.Add(stat.Task{
					Duration: time.Since(start),
				})
			}()
			return handler(ctx, req)
		}
	}
}
