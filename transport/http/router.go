package http

import (
	"context"
	"github.com/gin-gonic/gin"
	utils "github.com/go-slark/slark/pkg"
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

func (r *Router) Handle(method, path string, hf HandlerFunc) {
	handler := func(ctx *gin.Context) {
		mp := make(map[string]string, len(ctx.Params))
		for _, param := range ctx.Params {
			mp[param.Key] = param.Value
		}
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), utils.RequestVars, mp))
		c := r.pool.Get().(*Context)
		c.Set(ctx.Request, ctx.Writer)
		if err := hf(c); err != nil {
			r.srv.codecs.errorEncoder(ctx.Request, c.w, err)
		}
		c.Set(nil, nil)
		r.pool.Put(c)
	}
	r.srv.engine.Handle(method, r.srv.basePath+path, handler)
}
