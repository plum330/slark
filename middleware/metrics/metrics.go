package metrics

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/transport"
	"strconv"
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
		name:       "name",
		help:       "help",
		namespace:  "ns",
		subSystem:  "ss",
		labels:     []string{"kind", "operation", "code", "reason"},
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

func Help(h string) VecOpts {
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

var (
	RequestTotal = NewCounter(
		Namespace("client"),
		Name("code_count"),
		Help("client requests code count"),
		SubSystem("call"),
	)

	RequestDuration = NewHistogram(
		Namespace("server"),
		Name("duration_second"),
		Help("server requests duration second"),
		SubSystem("requests"),
		Labels([]string{"kind", "operation"}),
	)
)

func Metrics(st middleware.SubType, opts ...Options) middleware.Middleware {
	o := &Option{}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				kind, operation, reason string
				ok                      bool
				code                    int32
				trans                   transport.Transporter
			)
			if st == middleware.Client {
				trans, ok = transport.FromClientContext(ctx)
			} else if st == middleware.Server {
				trans, ok = transport.FromServerContext(ctx)
			}
			if !ok {
				return handler(ctx, req)
			}
			kind = trans.Kind()
			operation = trans.Operate()
			start := time.Now()
			rsp, err := handler(ctx, req)
			if err != nil {
				e := errors.FromError(err)
				reason = e.Reason
				code = e.Code
			}
			if o.histogram != nil {
				o.histogram.Values(kind, operation).Observe(time.Since(start).Seconds())
			}
			if o.counter != nil {
				o.counter.Values(kind, operation, strconv.Itoa(int(code)), reason).Inc()
			}
			return rsp, err
		}
	}
}
