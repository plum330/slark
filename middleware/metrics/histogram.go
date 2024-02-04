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

func NewHistogram(opt *VecOptions) Histogram {
	opts := prometheus.HistogramOpts{
		Namespace: opt.namespace,
		Subsystem: opt.subSystem,
		Name:      opt.name,
		Help:      opt.help,
	}
	vec := prometheus.NewHistogramVec(opts, opt.labels)
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
