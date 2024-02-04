package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// 跟踪服务调用关系
// tracer表示一次完整的链路跟踪,tracer由一个/多个span(链路跟踪基本组成要素,表示一次函数调用/http请求开始和节数，以及调用的服务名,方法名,参数,异常等)组成,将span发送到跟踪系统
// TracerProvider 用于创建tracer,一般是需要第三方的分布式链路跟踪管理平台提供具体的实现(zipkin,jaeger...),默认NoopTracerProvider虽然也能创建Tracer但是内部不会执行具体的数据流传输逻辑。
// TextMapPropagator传播器用于端对端数据编解码

type Tracing struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	kind       trace.SpanKind
	name       string
}

type Option func(option *Tracing)

func Name(name string) Option {
	return func(option *Tracing) {
		option.name = name
	}
}

func Provider(provider trace.TracerProvider) Option {
	return func(option *Tracing) {
		option.provider = provider
	}
}

func Propagator(propagator propagation.TextMapPropagator) Option {
	return func(option *Tracing) {
		option.propagator = propagator
	}
}

// propagation.Baggage保存链路跟踪过程中跨服务/进程的自定义k/v数据

func NewTracing(kind trace.SpanKind, opts ...Option) *Tracing {
	tracing := &Tracing{
		provider:   trace.NewNoopTracerProvider(),
		propagator: propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}),
		name:       "tracer",
		kind:       kind,
	}
	for _, opt := range opts {
		opt(tracing)
	}
	otel.SetTracerProvider(tracing.provider)
	tracing.tracer = otel.Tracer(tracing.name)
	return tracing
}

func (t *Tracing) Kind() trace.SpanKind {
	return t.kind
}

func (t *Tracing) Name() string {
	return t.name
}

func (t *Tracing) Start(ctx context.Context, name string, carrier propagation.TextMapCarrier, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.kind == trace.SpanKindServer {
		ctx = t.propagator.Extract(ctx, carrier)
	}
	// 创建span
	ctx, span := t.tracer.Start(ctx, name, opts...)
	if t.kind == trace.SpanKindClient {
		t.propagator.Inject(ctx, carrier)
	}
	return ctx, span
}

func (t *Tracing) Stop(span trace.Span, m interface{}, err error) {
	//if err != nil {
	//	span.RecordError(err)
	//	if e := errors.FromError(err); e != nil {
	//		span.SetAttributes(attribute.Key("rpc.status_code").Int64(int64(e.Code)))
	//	}
	//	span.SetStatus(codes.Error, err.Error())
	//} else {
	//	span.SetStatus(codes.Ok, "OK")
	//}
	//
	//if p, ok := m.(proto.Message); ok {
	//	if t.kind == trace.SpanKindServer {
	//		span.SetAttributes(attribute.Key("send_msg.size").Int(proto.Size(p)))
	//	} else {
	//		span.SetAttributes(attribute.Key("recv_msg.size").Int(proto.Size(p)))
	//	}
	//}
	span.End()
}
