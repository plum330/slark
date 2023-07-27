package trace

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type TracerOption struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
	name       string
}

type Option func(option *TracerOption)

func Name(name string) Option {
	return func(option *TracerOption) {
		option.name = name
	}
}

func Provider(provider trace.TracerProvider) Option {
	return func(option *TracerOption) {
		option.provider = provider
	}
}

func Propagator(propagator propagation.TextMapPropagator) Option {
	return func(option *TracerOption) {
		option.propagator = propagator
	}
}

type Tracer struct {
	tracer trace.Tracer
	opt    *TracerOption
	kind   trace.SpanKind
}

func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	o := &TracerOption{
		provider:   trace.NewNoopTracerProvider(),
		propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		name:       "slark",
	}
	for _, opt := range opts {
		opt(o)
	}
	otel.SetTracerProvider(o.provider)
	tracer := &Tracer{
		tracer: otel.Tracer(o.name),
		opt:    o,
		kind:   kind,
	}
	return tracer
}

func (t *Tracer) Start(ctx context.Context, spanName string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		ctx = t.opt.propagator.Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx, spanName, opts...)
	if t.kind == trace.SpanKindClient {
		t.opt.propagator.Inject(ctx, carrier)
	}
	return ctx, span
}

func (t *Tracer) Stop(_ context.Context, span trace.Span, m interface{}, err error) {
	if err != nil {
		span.RecordError(err)
		if e := errors.FromError(err); e != nil {
			span.SetAttributes(attribute.Key("rpc.status_code").Int64(int64(e.Code)))
		}
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	if p, ok := m.(proto.Message); ok {
		if t.kind == trace.SpanKindServer {
			span.SetAttributes(attribute.Key("send_msg.size").Int(proto.Size(p)))
		} else {
			span.SetAttributes(attribute.Key("recv_msg.size").Int(proto.Size(p)))
		}
	}
	span.End()
}

func HTTPServerTrace(opts ...Option) middleware.HTTPMiddleware {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opt := []trace.SpanStartOption{trace.WithSpanKind(tracer.kind), trace.WithAttributes(httpconv.ServerRequest(tracer.opt.name, r)...)}
			ctx, span := tracer.Start(r.Context(), r.URL.Path, propagation.HeaderCarrier(r.Header), opt...)
			handler.ServeHTTP(w, r.WithContext(ctx))
			span.End()
		})
	}
}

func GRPCServerTrace(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			name, _ := ctx.Value(utils.Method).(string)
			name, attr := SpanInfo(name, parseAddr(ctx))
			ctx, span := tracer.Start(ctx, name, &Metadata{&md}, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attr...))
			defer tracer.Stop(ctx, span, rsp, err)
			return handler(ctx, req)
		}
	}
}

func GRPCClientTrace(opts ...Option) middleware.Middleware {
	tracer := NewTracer(trace.SpanKindClient, opts...)
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
			for k, v := range m {
				name, attr = SpanInfo(k, v)
			}
			ctx, span := tracer.Start(ctx, name, &Metadata{&md}, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attr...))
			ctx = metadata.NewOutgoingContext(ctx, md)
			defer tracer.Stop(ctx, span, rsp, err)
			return handler(ctx, req)
		}
	}
}
