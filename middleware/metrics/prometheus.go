package metrics

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Log(context.TODO(), logger.FatalLevel, map[string]interface{}{"error": http.ListenAndServe(":8081", nil)}, "prometheus http")
	}()
}
