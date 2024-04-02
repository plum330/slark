package metrics

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/opentelemetry/metric"
	"github.com/go-slark/slark/transport"
	"go.opentelemetry.io/otel/attribute"
	"strconv"
	"time"
)

func Metrics(pt middleware.PeerType, opts ...metric.Option) middleware.Middleware {
	meter, err := metric.NewMeter(opts...)
	if err != nil {
		return func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, err
			}
		}
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				kind, operation, reason string
				ok                      bool
				code                    int32
				trans                   transport.Transporter
			)
			if pt == middleware.Client {
				trans, ok = transport.FromClientContext(ctx)
			} else if pt == middleware.Server {
				trans, ok = transport.FromServerContext(ctx)
			}
			if !ok {
				return handler(ctx, req)
			}
			kind = trans.Kind()
			operation = trans.Operate()
			start := time.Now()
			rsp, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				reason = e.Reason
				code = e.Code
			}
			meter.Histogram(ctx, time.Since(start).Milliseconds(),
				attribute.String("kind", kind),
				attribute.String("operation", operation),
			)
			meter.Counter(ctx,
				attribute.String("kind", kind),
				attribute.String("operation", operation),
				attribute.String("code", strconv.Itoa(int(code))),
				attribute.String("reason", reason),
			)
			return rsp, err
		}
	}
}
