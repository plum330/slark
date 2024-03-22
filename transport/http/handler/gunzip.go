package handler

import (
	"compress/gzip"
	utils "github.com/go-slark/slark/pkg"
	"net/http"
	"strings"
)

func Gunzip() Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get(utils.ContentEncoding), "gunzip") {
				reader, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = reader
			}
			handler.ServeHTTP(w, r)
		})
	}
}
