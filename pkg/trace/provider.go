package trace

import (
	"github.com/go-slark/slark/pkg/noop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
)

func init() {
	exporter, _ := stdouttrace.New(stdouttrace.WithWriter(noop.Writer()))
	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("slark"),
		)),
		sdktrace.WithBatcher(exporter),
	))
}

func NewZipkinProvider(url string) (trace.TracerProvider, error) {
	exporter, err := zipkin.New(
		url,
		zipkin.WithLogger(log.New(os.Stderr, "salrk", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("slark"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
