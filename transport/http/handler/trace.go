package handler

import (
	tracing "github.com/go-slark/slark/middleware/trace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.17.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func Trace(opts ...tracing.Option) Middleware {
	t := tracing.NewTracing(trace.SpanKindServer, opts...)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opt := []trace.SpanStartOption{
				trace.WithSpanKind(t.Kind()),
				trace.WithAttributes(httpconv.ServerRequest(t.Name(), r)...),
			}
			ctx, span := t.Start(r.Context(), r.URL.Path, propagation.HeaderCarrier(r.Header), opt...)
			handler.ServeHTTP(w, r.WithContext(ctx))
			span.End()
		})
	}
}
