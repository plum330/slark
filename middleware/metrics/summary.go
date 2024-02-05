package metrics

import "github.com/prometheus/client_golang/prometheus"

type Summary interface {
	Observe(float642 float64)
	Values(v ...string) Summary
}

type summary struct {
	*prometheus.SummaryVec
	values []string
}

func NewSummary(opts ...VecOpts) Summary {
	o := newVecOptions()
	for _, opt := range opts {
		opt(o)
	}
	sOpts := prometheus.SummaryOpts{
		Namespace:  o.namespace,
		Subsystem:  o.subSystem,
		Name:       o.name,
		Help:       o.help,
		Objectives: o.objectives,
	}
	vec := prometheus.NewSummaryVec(sOpts, o.labels)
	prometheus.MustRegister(vec)
	return &summary{
		SummaryVec: vec,
	}
}

func (s *summary) Observe(v float64) {
	s.WithLabelValues(s.values...).Observe(v)
}

func (s *summary) Values(v ...string) Summary {
	return &summary{
		SummaryVec: s.SummaryVec,
		values:     v,
	}
}
