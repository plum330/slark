package http

import (
	"github.com/gin-gonic/gin"
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
		return &Context{router: router}
	}
	return router
}

type HandlerFunc func(ctx *Context) error

func (r *Router) Handle(method, path string, hf HandlerFunc) {
	handler := func(ctx *gin.Context) {
		c := r.pool.Get().(*Context)
		c.Set(ctx.Request, ctx.Writer)
		if err := hf(c); err != nil {
			r.srv.Codecs.errorEncoder(ctx.Request, ctx.Writer, err)
		}
		c.Set(nil, nil)
		r.pool.Put(ctx)
	}
	r.srv.Engine.Handle(method, path, handler)
}
