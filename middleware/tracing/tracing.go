package tracing

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	tracing "github.com/go-slark/slark/pkg/trace"
	"github.com/go-slark/slark/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	"net/http"
)

func Trace(kind trace.SpanKind, opts ...tracing.Option) middleware.Middleware {
	tracer := tracing.NewTracer(kind, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				trans transport.Transporter
				ok    bool
			)
			if kind == trace.SpanKindClient {
				trans, ok = transport.FromClientContext(ctx)
			} else if kind == trace.SpanKindServer {
				trans, ok = transport.FromServerContext(ctx)
			}
			if !ok {
				return handler(ctx, req)
			}

			operation := trans.Operate()
			k := trans.Kind()
			var attrs []attribute.KeyValue
			if k == transport.GRPC {
				attrs = attributes(ctx, operation)
			} else if k == transport.HTTP {
				attrs = httpAttributes(operation)
			}
			opt := []trace.SpanStartOption{
				trace.WithSpanKind(kind),
				trace.WithAttributes(attrs...),
			}
			ctx, span := tracer.Start(ctx, operation, trans.ReqCarrier(), opt...)
			defer span.End()
			rsp, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				if k == transport.GRPC {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(e.Code)))
				} else if k == transport.HTTP {
					span.SetAttributes(semconv.HTTPStatusCode(int(e.Code)))
				}
				span.SetStatus(codes.Error, e.Reason)
			} else {
				if k == transport.GRPC {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(grpccodes.OK.String()))
				} else if k == transport.HTTP {
					span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
				}
			}
			return rsp, err
		}
	}
}
