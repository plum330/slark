package http

import (
	"context"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"net/http"
)

type wrapper struct {
	rw   http.ResponseWriter
	code int
	err  error // propagate error for middleware
}

func (w *wrapper) WriteHeader(code int) {
	w.code = code
	w.rw.WriteHeader(code)
}

func (w *wrapper) Header() http.Header {
	return w.rw.Header()
}

func (w *wrapper) Write(p []byte) (int, error) {
	return w.rw.Write(p)
}

type Context struct {
	router *Router
	req    *http.Request
	rsp    http.ResponseWriter
	ctx    context.Context
	w      *wrapper
}

func (c *Context) Set(req *http.Request, rsp http.ResponseWriter) {
	c.req = req
	c.rsp = rsp
	c.w = &wrapper{
		rw: rsp,
	}
	if c.req == nil || rsp == nil {
		return
	}
	c.ctx = c.req.Context()
	for _, hk := range c.router.srv.headers {
		hv := c.req.Header.Get(hk)
		if len(hv) == 0 {
			continue
		}
		c.ctx = context.WithValue(c.ctx, hk, hv)
	}
	c.ctx = context.WithValue(c.ctx, utils.Method, req.Method)
	c.ctx = context.WithValue(c.ctx, utils.Path, req.URL.Path)
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Handle(handler middleware.Handler) middleware.Handler {
	return middleware.ComposeMiddleware(c.router.srv.mws...)(handler)
}

func (c *Context) ShouldBind(v interface{}) error {
	return c.router.srv.Codecs.bodyDecoder(c.req, v)
}

func (c *Context) ShouldBindURI(v interface{}) error {
	return c.router.srv.Codecs.varsDecoder(c.req, v)
}

func (c *Context) ShouldBindQuery(v interface{}) error {
	return c.router.srv.Codecs.queryDecoder(c.req, v)
}

func (c *Context) Result(v interface{}) error {
	return c.router.srv.Codecs.rspEncoder(c.req, c.w, v)
}
