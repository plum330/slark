package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCounter(t *testing.T) {
	c := NewCounter(prometheus.CounterOpts{
		Namespace: "http_server",
		Subsystem: "requests",
		Name:      "total",
		Help:      "http server requests count",
	}, []string{"url", "method"})
	c.Values("/server/api", "POST").Inc()
	c.Values("/server/api", "POST").Add(2)
	assert.NotNil(t, c)
	assert.Equal(t, float64(3), testutil.ToFloat64(c.(*counter)))
}

func TestGauge(t *testing.T) {
	g := NewGauge(prometheus.GaugeOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "duration",
		Help:      "server requests duration",
	}, []string{"url"})
	assert.NotNil(t, g)
	g.Values("/server/api").Inc()
	g.Values("/server/api").Add(1)
	assert.Equal(t, float64(2), testutil.ToFloat64(g.(*gauge)))
}

func TestHistogram(t *testing.T) {
	h := NewHistogram(prometheus.HistogramOpts{
		Name:    "duration_ms",
		Help:    "server requests duration",
		Buckets: []float64{1, 2, 3},
	}, []string{"url"})
	assert.NotNil(t, h)
	h.Values("/server/api").Observe(1)
	metadata := `
		# HELP counts rpc server requests duration(ms).
        # TYPE counts histogram
`
	val := `
		counts_bucket{method="/server/api",le="1"} 0
		counts_bucket{method="/server/api",le="2"} 1
		counts_bucket{method="/server/api",le="3"} 1
		counts_bucket{method="/server/api",le="+Inf"} 1
		counts_sum{method="/server/api"} 2
        counts_count{method="/server/api"} 1
`
	err := testutil.CollectAndCompare(h.(*histogram), strings.NewReader(metadata+val))
	assert.Nil(t, err)
}

/*
var (
	rpcServerSeconds = NewHistogram(prometheus.HistogramOpts{
		Namespace: "rpc-server",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc server requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"method"})

	rpcServerRequests = NewCounter(prometheus.CounterOpts{
		Namespace: "rpc-server",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "The total number of processed requests",
	}, []string{"method", "code"})

	rpcClientSeconds = NewHistogram(prometheus.HistogramOpts{
		Namespace: "rpc-client",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc client requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"method"})

	rpcClientRequests = NewCounter(prometheus.CounterOpts{
		Namespace: "rpc-client",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc client requests code count.",
	}, []string{"method", "code"})

	httpServerSeconds = NewHistogram(prometheus.HistogramOpts{
		Namespace: "http-server",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"path", "method"})

	httpServerRequests = NewCounter(prometheus.CounterOpts{
		Namespace: "http-server",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests count.",
	}, []string{"path", "method"})
)
*/
