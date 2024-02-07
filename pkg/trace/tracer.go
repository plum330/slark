package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

// propagation.Baggage保存链路跟踪过程中跨服务/进程的自定义k/v数据

func NewTracer(kind trace.SpanKind, opts ...Option) *Tracer {
	tracer := &Tracer{
		provider:   otel.GetTracerProvider(),
		propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		kind:       kind,
	}
	for _, opt := range opts {
		opt(tracer)
	}
	otel.SetTracerProvider(tracer.provider)
	tracer.tracer = otel.Tracer(tracer.name)
	return tracer
}

func (t *Tracer) Start(ctx context.Context, name string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		// 将carrier从metadata中提取出来，创建span,如此client端与server端就能建立span信息的关联
		ctx = t.propagator.Extract(ctx, carrier)
	}
	// 创建span
	ctx, span := t.tracer.Start(ctx, name, opts...)
	if t.kind == trace.SpanKindClient {
		// 将span的context信息注入到carrier，再将carrier写入到metadata中，完成span信息的传递
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
