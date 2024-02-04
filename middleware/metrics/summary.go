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

func NewSummary(opt *VecOptions) Summary {
	opts := prometheus.SummaryOpts{
		Namespace:  opt.namespace,
		Subsystem:  opt.subSystem,
		Name:       opt.name,
		Help:       opt.help,
		Objectives: opt.objectives,
	}
	vec := prometheus.NewSummaryVec(opts, opt.labels)
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
