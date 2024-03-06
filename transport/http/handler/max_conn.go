package handler

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/limit"
	"net/http"
)

func MaxConn(l logger.Logger, n int) Middleware {
	if n <= 0 {
		return func(handler http.Handler) http.Handler {
			return handler
		}
	}
	pool := limit.NewPool(n)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allow := pool.Use()
			if !allow {
				w.WriteHeader(http.StatusServiceUnavailable)
				l.Log(context.TODO(), logger.WarnLevel, map[string]interface{}{"req": fmt.Sprintf("%+v", r)}, "conn overload")
				return
			}
			defer func() {
				err := pool.Back()
				if err != nil {
					l.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err})
				}
			}()
			handler.ServeHTTP(w, r)
		})
	}
}
