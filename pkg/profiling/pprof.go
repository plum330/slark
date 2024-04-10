package profiling

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/zeromicro/go-zero/core/proc"
	"net/http"
	_ "net/http/pprof"
)

// kill -usr1 pid

// kill -usr2 pid

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Log(context.TODO(), logger.FatalLevel, map[string]interface{}{"error": http.ListenAndServe(":8081", nil)})
	}()
}
