package metrics

import "github.com/prometheus/client_golang/prometheus"

type Gauge interface {
	Set(float64)
	Inc()
	Add(v float64)
	Values(v ...string) Gauge
}

type gauge struct {
	*prometheus.GaugeVec
	values []string
}

func NewGauge(opts ...VecOpts) Gauge {
	o := newVecOptions()
	for _, opt := range opts {
		opt(o)
	}
	gOpts := prometheus.GaugeOpts{
		Namespace: o.namespace,
		Subsystem: o.subSystem,
		Name:      o.name,
		Help:      o.help,
	}
	vec := prometheus.NewGaugeVec(gOpts, o.labels)
	prometheus.MustRegister(vec)
	return &gauge{
		GaugeVec: vec,
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

func (g *gauge) Values(v ...string) Gauge {
	return &gauge{
		GaugeVec: g.GaugeVec,
		values:   v,
	}
}
