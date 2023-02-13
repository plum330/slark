package metrics

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

type Counter interface {
	Inc()
	Add(float64)
	Values(...string) Counter
}

type counter struct {
	*prometheus.CounterVec
	values []string
}

func NewCounter(opts prometheus.CounterOpts, labels []string) Counter {
	return &counter{
		CounterVec: prometheus.NewCounterVec(opts, labels),
	}
}

func (c *counter) Inc() {
	c.WithLabelValues(c.values...).Inc()
}

func (c *counter) Add(add float64) {
	c.WithLabelValues(c.values...).Add(add)
}

func (c *counter) Values(values ...string) Counter {
	return &counter{
		CounterVec: c.CounterVec,
		values:     values,
	}
}

// gauge

type Gauge interface {
	Set(float64)
	Inc()
	Add(v float64)
	Values(...string) Gauge
}

type gauge struct {
	*prometheus.GaugeVec
	values []string
}

func NewGauge(opts prometheus.GaugeOpts, labels []string) Gauge {
	return &gauge{
		GaugeVec: prometheus.NewGaugeVec(opts, labels),
	}
}

func (g *gauge) Set(v float64) {
	g.WithLabelValues(g.values...).Set(v)
}

func (g *gauge) Inc() {
	g.WithLabelValues(g.values...).Inc()
}

func (g *gauge) Add(v float64) {
	g.WithLabelValues(g.values...).Add(v)
}

func (g *gauge) Values(values ...string) Gauge {
	return &gauge{
		GaugeVec: g.GaugeVec,
		values:   values,
	}
}

// histogram

type Histogram interface {
	Observe(float64)
	Values(...string) Histogram
}

type histogram struct {
	*prometheus.HistogramVec
	labelValues []string
}

func NewHistogram(opts prometheus.HistogramOpts, labels []string) Histogram {
	return &histogram{
		HistogramVec: prometheus.NewHistogramVec(opts, labels),
	}
}

func (h *histogram) Observe(v float64) {
	h.WithLabelValues(h.labelValues...).Observe(v)
}

func (h *histogram) Values(values ...string) Histogram {
	return &histogram{
		HistogramVec: h.HistogramVec,
		labelValues:  values,
	}
}

// mw

type option struct {
	counter   Counter
	gauge     Gauge
	histogram Histogram
}

type Options func(*option)

func WithCounter(c Counter) Options {
	return func(o *option) {
		o.counter = c
	}
}

func WithGauge(g Gauge) Options {
	return func(o *option) {
		o.gauge = g
	}
}

func WithHistogram(h Histogram) Options {
	return func(o *option) {
		o.histogram = h
	}
}

func HTTPMetrics(opts ...Options) func(http.Handler) http.Handler {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tm := time.Now()
			handler.ServeHTTP(w, r)
			if o.counter != nil {
				o.counter.Values(r.URL.Path, r.Method).Inc()
			}
			if o.histogram != nil {
				o.histogram.Values(r.URL.Path, r.Method).Observe(time.Since(tm).Seconds())
			}
		})
	}
}

func GRPCMetrics(opts ...Options) middleware.Middleware {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tm := time.Now()
			rsp, err := handler(ctx, req)
			method, _ := ctx.Value(struct{}{}).(string)
			if o.counter != nil {
				o.counter.Values(method).Inc()
			}

			if o.histogram != nil {
				o.histogram.Values(method).Observe(time.Since(tm).Seconds())
			}
			return rsp, err
		}
	}
}
