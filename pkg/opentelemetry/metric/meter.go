package metric

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Meter struct {
	name      string
	provider  metric.MeterProvider
	meter     metric.Meter
	counter   metric.Int64Counter
	histogram metric.Float64Histogram
}

type Option func(*Meter)

func WithCounter(counter metric.Int64Counter) Option {
	return func(m *Meter) {
		m.counter = counter
	}
}

func WithHistogram(histogram metric.Float64Histogram) Option {
	return func(m *Meter) {
		m.histogram = histogram
	}
}

func WithProvider(provider metric.MeterProvider) Option {
	return func(m *Meter) {
		m.provider = provider
	}
}

func WithName(name string) Option {
	return func(m *Meter) {
		m.name = name
	}
}

func RequestCodeCounter() metric.Int64Counter {
	m := otel.Meter("slark")
	counter, _ := m.Int64Counter("code_count")
	return counter
}

func RequestDurationHistogram() metric.Float64Histogram {
	m := otel.Meter("slark")
	his, _ := m.Float64Histogram("duration_seconds", metric.WithUnit("s"))
	return his
}

func NewMeter(opts ...Option) *Meter {
	m := &Meter{
		name: "slark",
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.provider == nil {
		m.provider = otel.GetMeterProvider()
	}
	m.meter = m.provider.Meter(m.name)
	return m
}

func (m *Meter) Counter(ctx context.Context, attributes ...attribute.KeyValue) {
	if m.counter != nil {
		m.counter.Add(ctx, 1, metric.WithAttributes(attributes...))
	}
}

func (m *Meter) Histogram(ctx context.Context, incr float64, attributes ...attribute.KeyValue) {
	if m.histogram != nil {
		m.histogram.Record(ctx, incr, metric.WithAttributes(attributes...))
	}
}
