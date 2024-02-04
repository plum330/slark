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

func NewCounter(opt *VecOptions) Counter {
	opts := prometheus.CounterOpts{
		Namespace: opt.namespace,
		Subsystem: opt.subSystem,
		Name:      opt.name,
		Help:      opt.help,
	}
	vec := prometheus.NewCounterVec(opts, opt.labels)
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
