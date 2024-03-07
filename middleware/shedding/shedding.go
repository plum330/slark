package shedding

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/zeromicro/go-zero/core/load"
)

func Shedding(kind string, threshold int64) middleware.Middleware {
	load.DisableLog()
	sheddingStat := load.NewSheddingStat(kind)
	shedding := load.NewAdaptiveShedder(load.WithCpuThreshold(threshold))
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			sheddingStat.IncrementTotal()
			promise, err := shedding.Allow()
			if err != nil {
				sheddingStat.IncrementDrop()
				err = errors.ServerUnavailable(err.Error(), err.Error())
				return nil, err
			}
			defer func() {
				if errors.Is(err, context.DeadlineExceeded) || errors.IsServerUnavailable(err) {
					promise.Fail()
				} else {
					sheddingStat.IncrementPass()
					promise.Pass()
				}
			}()
			return handler(ctx, req)
		}
	}
}
