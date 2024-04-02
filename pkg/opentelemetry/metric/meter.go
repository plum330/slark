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
	histogram metric.Int64Histogram
}

type Option func(*Meter)

func WithCounter(counter metric.Int64Counter) Option {
	return func(m *Meter) {
		m.counter = counter
	}
}

func WithHistogram(histogram metric.Int64Histogram) Option {
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

func NewMeter(opts ...Option) (*Meter, error) {
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
	if m.counter == nil {
		counter, err := m.meter.Int64Counter("code_count")
		if err != nil {
			return nil, err
		}
		m.counter = counter
	}
	if m.histogram == nil {
		histogram, err := m.meter.Int64Histogram("duration_second", metric.WithUnit("s"))
		if err != nil {
			return nil, err
		}
		m.histogram = histogram
	}
	return m, nil
}

func (m *Meter) Counter(ctx context.Context, attributes ...attribute.KeyValue) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attributes...))
}

func (m *Meter) Histogram(ctx context.Context, incr int64, attributes ...attribute.KeyValue) {
	m.histogram.Record(ctx, incr, metric.WithAttributes(attributes...))
}
