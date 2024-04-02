package trace

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

func ExtractTraceID(ctx context.Context) string {
	var traceID string
	sCtx := trace.SpanContextFromContext(ctx)
	if sCtx.HasTraceID() {
		traceID = sCtx.TraceID().String()
	}
	return traceID
}

func ExtractSpanID(ctx context.Context) string {
	var spanID string
	sCtx := trace.SpanContextFromContext(ctx)
	if sCtx.HasSpanID() {
		spanID = sCtx.SpanID().String()
	}
	return spanID
}
