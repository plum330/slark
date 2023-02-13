package trace

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
	"os"
	"testing"
)

// stdout
func TestStdoutTrace(t *testing.T) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		fmt.Printf("creating stdout exporter: %+v\n", err)
		return
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("stdout-example"),
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	defer tp.Shutdown(context.TODO())
	HTTPServerTrace(Provider(tp))
}

// jaeger

func TestJaegerTrace(t *testing.T) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(""))) // jaeger server url
	if err != nil {
		fmt.Printf("jaeger init err:%+v\n", err)
		return
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(1.0))),
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String("slark"),
			attribute.String("env", "dev"),
		)),
	)
	defer tp.Shutdown(context.TODO())
	HTTPServerTrace(Provider(tp))
}

// zipkin in mem

func TestZipkinTrace(t *testing.T) {
	exporter, err := zipkin.New(
		"http://127.0.0.1:9411/api/v2/spans", // url
		zipkin.WithLogger(log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		fmt.Printf("zipkin init err:%+v\n", err)
		return
	}

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exporter)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("zipkin-test"),
		)),
	)
	defer tp.Shutdown(context.TODO())
	HTTPServerTrace(Provider(tp))
}
