package handler

import (
	"github.com/go-slark/slark/logger"
	"net/http"
)

func MaxBytes(n int64) Middleware {
	if n <= 0 {
		return func(handler http.Handler) http.Handler {
			return handler
		}
	}
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				logger.Log(r.Context(), logger.ErrorLevel, map[string]interface{}{"length": r.ContentLength, "limit": n}, "request entity too large")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
			} else {
				handler.ServeHTTP(w, r)
			}
		})
	}
}
