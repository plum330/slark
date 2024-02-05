package metrics

import "github.com/prometheus/client_golang/prometheus"

type Histogram interface {
	Observe(float64)
	Values(v ...string) Histogram
}

type histogram struct {
	*prometheus.HistogramVec
	values []string
}

func NewHistogram(opts ...VecOpts) Histogram {
	o := newVecOptions()
	for _, opt := range opts {
		opt(o)
	}
	hOpts := prometheus.HistogramOpts{
		Namespace: o.namespace,
		Subsystem: o.subSystem,
		Name:      o.name,
		Help:      o.help,
		Buckets:   o.buckets,
	}
	vec := prometheus.NewHistogramVec(hOpts, o.labels)
	prometheus.MustRegister(vec)
	return &histogram{
		HistogramVec: vec,
	}
}

func (h *histogram) Observe(v float64) {
	h.WithLabelValues(h.values...).Observe(v)
}

func (h *histogram) Values(v ...string) Histogram {
	return &histogram{
		HistogramVec: h.HistogramVec,
		values:       v,
	}
}
