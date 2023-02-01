package middleware

import (
	"github.com/go-slark/slark/pkg"
	"github.com/rs/cors"
	"net/http"
)

type CORSOption func(options *cors.Options)

func AllowCredentials(allow bool) CORSOption {
	return func(options *cors.Options) {
		options.AllowCredentials = allow
	}
}

func AllowedMethods(methods []string) CORSOption {
	return func(options *cors.Options) {
		options.AllowedMethods = methods
	}
}

func AllowOriginFunc(allow bool) CORSOption {
	return func(options *cors.Options) {
		options.AllowOriginFunc = func(origin string) bool {
			return allow
		}
	}
}

func AllowedHeaders(headers []string) CORSOption {
	return func(options *cors.Options) {
		options.AllowedHeaders = headers
	}
}

func MaxAge(age int) CORSOption {
	return func(options *cors.Options) {
		options.MaxAge = age
	}
}

func CORS(opts ...CORSOption) HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			o := cors.Options{
				AllowCredentials: true,
				AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
				AllowOriginFunc:  func(origin string) bool { return true },
				AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type", "Accept-Encoding", "Authorization", "X-CSRF-Token", pkg.Authorization, "Content-Disposition"},
				ExposedHeaders:   []string{pkg.Authorization, "Content-Disposition"},
				MaxAge:           43200, // 12 Hours
			}
			for _, opt := range opts {
				opt(&o)
			}
			cors.New(o).HandlerFunc(writer, request)
			handler.ServeHTTP(writer, request)
		}
	}
}
