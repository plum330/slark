package trace

import (
	"context"
	utils "github.com/go-slark/slark/pkg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// 初始化后，其他地方可以直接获取全局的tracer使用
// 将请求与响应的详细信息全都记录到 span 中，如 URL、Method、请求头、请求体、响应状态码、响应头、响应体等

type Tracer struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	kind       trace.SpanKind
	name       string
}

type Option func(option *Tracer)

func Name(name string) Option {
	return func(option *Tracer) {
		option.name = name
	}
}

func Provider(provider trace.TracerProvider) Option {
	return func(option *Tracer) {
		option.provider = provider
	}
}

func Propagator(propagator propagation.TextMapPropagator) Option {
	return func(option *Tracer) {
		option.propagator = propagator
	}
}

func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	exporter, _ := stdouttrace.New(stdouttrace.WithWriter(utils.NoopWriter()))
	tracer := &Tracer{
		provider: sdktrace.NewTracerProvider(
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("slark"),
			)),
			sdktrace.WithBatcher(exporter),
		),
		propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		kind:       kind,
		name:       "slark",
	}
	for _, opt := range opts {
		opt(tracer)
	}
	otel.SetTracerProvider(tracer.provider)
	tracer.tracer = otel.Tracer(tracer.name)
	return tracer
}

func (t *Tracer) Start(ctx context.Context, name string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer || t.kind == trace.SpanKindConsumer {
		ctx = t.propagator.Extract(ctx, carrier)
	}
	ctx, span := t.tracer.Start(ctx, name, opts...)
	if t.kind == trace.SpanKindClient || t.kind == trace.SpanKindProducer {
		t.propagator.Inject(ctx, carrier)
	}
	return ctx, span
}

func (t *Tracer) Kind() trace.SpanKind {
	return t.kind
}

func (t *Tracer) Name() string {
	return t.name
}
