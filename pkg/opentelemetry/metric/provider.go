package metric

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"net/http"
)

func provider(reader sdkmetric.Reader) metric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(func(instrument sdkmetric.Instrument) (sdkmetric.Stream, bool) {
			// counter
			return sdkmetric.Stream{
				Name:        instrument.Name,
				Description: instrument.Description,
				Unit:        instrument.Unit,
				AttributeFilter: func(value attribute.KeyValue) bool {
					return true
				},
			}, true
			// bucket
		}, func(instrument sdkmetric.Instrument) (sdkmetric.Stream, bool) {
			stream := sdkmetric.Stream{
				Name:        instrument.Name,
				Description: instrument.Description,
				Unit:        instrument.Unit,
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 1}, // s
					NoMinMax:   true,
				},
			}
			return stream, true
		}))
}

//func init() {
//	exporter, _ := stdoutmetric.New(stdoutmetric.WithWriter(noop.Writer()))
//	// NewPeriodicReader 设置metrics发送间隔
//	otel.SetMeterProvider(provider(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(500*time.Millisecond))))
//}

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Log(context.TODO(), logger.FatalLevel, map[string]interface{}{"error": http.ListenAndServe(":8081", nil)})
	}()
}

func init() {
	exporter, _ := prometheus.New()
	mp := provider(exporter)
	otel.SetMeterProvider(mp)
}
