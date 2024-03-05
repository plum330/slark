package tracing

import (
	"context"
	"github.com/go-slark/slark/middleware"
	tracing "github.com/go-slark/slark/pkg/trace"
	"github.com/go-slark/slark/transport"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Trace(kind trace.SpanKind, opts ...tracing.Option) middleware.Middleware {
	tracer := tracing.NewTracer(kind, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				trans transport.Transporter
				ok    bool
				md    metadata.MD
			)
			kind := tracer.Kind()
			if kind == trace.SpanKindClient {
				md, ok = metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.MD{}
				}
				trans, ok = transport.FromClientContext(ctx)
			} else {
				md, ok = metadata.FromIncomingContext(ctx)
				if !ok {
					md = metadata.MD{}
				}
				trans, ok = transport.FromServerContext(ctx)
			}
			if !ok {
				return handler(ctx, req)
			}
			name, attrs := attribute(ctx, trans.Operate())
			opt := []trace.SpanStartOption{
				trace.WithSpanKind(tracer.Kind()),
				trace.WithAttributes(attrs...),
			}
			ctx, span := tracer.Start(ctx, name, &tracing.Carrier{MD: &md}, opt...)
			if kind == trace.SpanKindClient {
				ctx = metadata.NewOutgoingContext(ctx, md)
			} else {
				ctx = metadata.NewIncomingContext(ctx, md)
			}
			defer span.End()
			rsp, err := handler(ctx, req)
			if err != nil {
				s, _ := status.FromError(err)
				span.SetStatus(codes.Error, s.Message())
				span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(s.Code().String()))
			} else {
				span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(grpccodes.OK.String()))
			}
			return rsp, err
		}
	}
}
