package breaker

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	bre "github.com/go-slark/slark/pkg/breaker"
	"strings"
)

// Breaker client in
func Breaker(opts ...bre.Option) middleware.Middleware {
	breakers := bre.NewBreaker(opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			str := []string{ctx.Value(utils.Method).(string), ctx.Value(utils.Path).(string)}
			breaker := breakers.Fetch(strings.Join(str, " "))
			err := breaker.Allow()
			if err != nil {
				breaker.Fail()
				return nil, errors.ServerUnavailable("trigger breaker", "TRIGGER_BREAKER")
			}
			rsp, err := handler(ctx, req)
			if err != nil {
				if errors.IsServerUnavailable(err) || errors.IsInternalServer(err) || errors.IsServerTimeout(err) {
					breaker.Fail()
				}
			} else {
				breaker.Succeed()
			}
			return rsp, err
		}
	}
}
