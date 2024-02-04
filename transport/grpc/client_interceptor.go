package grpc

import (
	"context"
	"github.com/go-slark/slark/middleware"
	tracing "github.com/go-slark/slark/middleware/trace"
	utils "github.com/go-slark/slark/pkg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

func (o *option) interceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = context.WithValue(ctx, utils.Target, map[string]string{method: cc.Target()})
		_, err := middleware.ComposeMiddleware(o.mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		})(ctx, req)
		return err
	}
}

func UnaryClientTimeout(time time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok && time > 0 {
			ctx, cancel = context.WithTimeout(ctx, time)
		}
		if cancel != nil {
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientTrace() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		value := ctx.Value(utils.RayID)
		requestID, ok := value.(string)
		if !ok || len(requestID) == 0 {
			requestID = utils.BuildRequestID()
		}
		ctx = metadata.AppendToOutgoingContext(ctx, utils.RayID, requestID)
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

func UnaryClientAuthorize() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token, ok := ctx.Value(utils.Token).(string)
		if ok {
			ctx = metadata.AppendToOutgoingContext(ctx, utils.Token, strconv.QuoteToASCII(token))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func ClientTrace(opts ...tracing.Option) middleware.Middleware {
	tracer := tracing.NewTracing(trace.SpanKindClient, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			m, _ := ctx.Value(utils.Target).(map[string]string)
			var (
				name string
				attr []attribute.KeyValue
			)
			for _, _ = range m {
				//name, attr = SpanInfo(k, v)
			}
			ctx, span := tracer.Start(ctx, name, &tracing.Carrier{MD: md}, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attr...))
			ctx = metadata.NewOutgoingContext(ctx, md)
			//defer tracer.Stop(ctx, span, rsp, err)
			span.End()
			return handler(ctx, req)
		}
	}
}
