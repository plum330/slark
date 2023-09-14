package cors

import (
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"github.com/rs/cors"
	"net/http"
)

type Option func(options *cors.Options)

func AllowCredentials(allow bool) Option {
	return func(options *cors.Options) {
		options.AllowCredentials = allow
	}
}

func AllowedMethods(methods []string) Option {
	return func(options *cors.Options) {
		options.AllowedMethods = methods
	}
}

func AllowOriginFunc(allow bool) Option {
	return func(options *cors.Options) {
		options.AllowOriginFunc = func(origin string) bool {
			return allow
		}
	}
}

func AllowOrigins(origins []string) Option {
	return func(options *cors.Options) {
		options.AllowedOrigins = origins
	}
}

func AllowedHeaders(headers []string) Option {
	return func(options *cors.Options) {
		options.AllowedHeaders = headers
	}
}

func ExposedHeaders(headers []string) Option {
	return func(options *cors.Options) {
		options.ExposedHeaders = headers
	}
}

func MaxAge(age int) Option {
	return func(options *cors.Options) {
		options.MaxAge = age
	}
}

func CORS(opts ...Option) middleware.HTTPMiddleware {
	return func(handler http.Handler) http.Handler {
		options := cors.Options{
			AllowCredentials: true,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions},
			AllowOriginFunc:  func(origin string) bool { return true },
			AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type", "Accept-Encoding", "Authorization", "X-CSRF-Token", utils.Authorization, utils.Token, "Content-Disposition"},
			ExposedHeaders:   []string{utils.Authorization, utils.Token, "Content-Disposition"},
			MaxAge:           43200, // 12 Hours
		}
		for _, opt := range opts {
			opt(&options)
		}
		return cors.New(options).Handler(handler)
	}
}
