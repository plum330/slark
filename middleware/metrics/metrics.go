package metrics

import (
	"context"
	"github.com/go-slark/slark/middleware"
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
	objectives map[float64]float64
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
	o := &Option{}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			rsp, err := handler(ctx, req)
			// path method from ctx, code from error
			if o.histogram != nil {
				o.histogram.Values().Observe(float64(time.Since(start).Milliseconds()))
			}
			if o.counter != nil {
				o.counter.Values().Inc()
			}
			return rsp, err
		}
	}
}
