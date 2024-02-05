package metrics

import "github.com/prometheus/client_golang/prometheus"

type Counter interface {
	Inc()
	Add(float64)
	Values(v ...string) Counter
}

type counter struct {
	*prometheus.CounterVec
	values []string
}

func NewCounter(opts ...VecOpts) Counter {
	o := newVecOptions()
	for _, opt := range opts {
		opt(o)
	}
	cOpts := prometheus.CounterOpts{
		Namespace: o.namespace,
		Subsystem: o.subSystem,
		Name:      o.name,
		Help:      o.help,
	}
	vec := prometheus.NewCounterVec(cOpts, o.labels)
	prometheus.MustRegister(vec)
	return &counter{
		CounterVec: vec,
	}
}

func (c *counter) Inc() {
	c.WithLabelValues(c.values...).Inc()
}

func (c *counter) Add(add float64) {
	c.WithLabelValues(c.values...).Add(add)
}

func (c *counter) Values(v ...string) Counter {
	return &counter{
		CounterVec: c.CounterVec,
		values:     v,
	}
}
