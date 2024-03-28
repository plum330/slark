package http

import (
	"context"
	"github.com/gin-gonic/gin"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/transport/http/handler"
	"net/http"
	"sync"
)

type Router struct {
	pool sync.Pool
	srv  *Server
}

func NewRouter(srv *Server) *Router {
	router := &Router{
		srv: srv,
	}
	router.pool.New = func() any {
		return &Context{
			router: router,
			ctx:    context.Background(),
		}
	}
	return router
}

type HandlerFunc func(ctx *Context) error

func (r *Router) Handle(method, path string, hf HandlerFunc, handlers ...handler.Middleware) {
	h := func(ctx *gin.Context) {
		// /uri/:name/:id
		mp := make(map[string]string, len(ctx.Params))
		for _, param := range ctx.Params {
			mp[param.Key] = param.Value
		}
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), utils.RequestVars, mp))
		handler.ComposeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := r.pool.Get().(*Context)
			c.Set(req, w)
			if err := hf(c); err != nil {
				r.srv.codecs.errorEncoder(req, w, err)
			}
			c.Set(nil, nil)
			r.pool.Put(c)
		}), handlers...).ServeHTTP(ctx.Writer, ctx.Request)
	}
	r.srv.engine.Handle(method, r.srv.basePath+path, h)
}
