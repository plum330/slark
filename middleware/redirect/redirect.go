package redirect

import (
	"github.com/go-slark/slark/middleware"
	"net/http"
)

type Redirecting struct {
	URL         string
	RedirectURL string
	Code        int
}

func Redirect(redirect *Redirecting) middleware.HTTPMiddleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == redirect.URL {
				http.Redirect(w, r, redirect.RedirectURL, redirect.Code)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}
