package metrics

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

/*
constant label 常量/固定标签对应的value不可变
非固定标签对应的value可通过WithLabelValues改变
labels顺序和values顺序对应
*/

type VecOptions struct {
	name       string
	help       string
	namespace  string
	subSystem  string
	labels     []string
	buckets    []float64
	objectives map[float64]float64
}

func newVecOptions() *VecOptions {
	return &VecOptions{
		name:       "vec",
		help:       "help",
		namespace:  "server",
		subSystem:  "request",
		labels:     []string{"method", "path", "code"},
		buckets:    []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1},
		objectives: nil,
	}
}

type VecOpts func(options *VecOptions)

func Name(name string) VecOpts {
	return func(o *VecOptions) {
		o.name = name
	}
}

func Namespace(ns string) VecOpts {
	return func(o *VecOptions) {
		o.namespace = ns
	}
}

func help(h string) VecOpts {
	return func(o *VecOptions) {
		o.help = h
	}
}

func SubSystem(s string) VecOpts {
	return func(o *VecOptions) {
		o.subSystem = s
	}
}

func Labels(labels []string) VecOpts {
	return func(o *VecOptions) {
		o.labels = labels
	}
}

func Buckets(buckets []float64) VecOpts {
	return func(o *VecOptions) {
		o.buckets = buckets
	}
}

func Objectives(objectives map[float64]float64) VecOpts {
	return func(o *VecOptions) {
		o.objectives = objectives
	}
}

type Option struct {
	counter   Counter
	gauge     Gauge
	histogram Histogram
}

type Options func(*Option)

func WithCounter(c Counter) Options {
	return func(o *Option) {
		o.counter = c
	}
}

func WithGauge(g Gauge) Options {
	return func(o *Option) {
		o.gauge = g
	}
}

func WithHistogram(h Histogram) Options {
	return func(o *Option) {
		o.histogram = h
	}
}

func Metrics(opts ...Options) middleware.Middleware {
	o := &Option{
		counter:   NewCounter(),
		histogram: NewHistogram(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var kind, operation string
			trans, ok := transport.FromServerContext(ctx)
			if ok {
				kind = trans.Kind()
				operation = trans.Operate()
			}
			start := time.Now()
			rsp, err := handler(ctx, req)
			if o.histogram != nil {
				o.histogram.Values(kind, operation).Observe(float64(time.Since(start).Milliseconds()))
			}
			if o.counter != nil {
				o.counter.Values(kind, operation).Inc()
			}
			return rsp, err
		}
	}
}

func init() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8081", nil)
}
