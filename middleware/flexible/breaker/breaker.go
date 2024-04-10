package breaker

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	bre "github.com/go-slark/slark/pkg/flexible/breaker"
	"github.com/go-slark/slark/transport"
)

func Breaker(opts ...bre.Option) middleware.Middleware {
	breakers := bre.NewBreaker(opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			trans, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			breaker := breakers.Fetch(trans.Operate())
			err := breaker.Allow()
			if err != nil {
				breaker.Fail(err.Error())
				return nil, errors.ServerUnavailable("trigger breaker", err.Error())
			}
			rsp, err := handler(ctx, req)
			if err != nil {
				if errors.IsServerUnavailable(err) || errors.IsInternalServer(err) || errors.IsServerTimeout(err) {
					breaker.Fail(err.Error())
				}
			} else {
				breaker.Succeed()
			}
			return rsp, err
		}
	}
}
