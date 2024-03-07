package http

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/transport/http/handler"
	"net/http"
)

type Context struct {
	router *Router
	req    *http.Request
	rsp    http.ResponseWriter
	ctx    context.Context
	w      *handler.Wrapper
}

func (c *Context) Set(req *http.Request, rsp http.ResponseWriter) {
	c.req = req
	c.rsp = rsp
	w := &handler.Wrapper{}
	w.SetResponseWriter(rsp)
	c.w = w
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Handle(handler middleware.Handler) middleware.Handler {
	return middleware.ComposeMiddleware(c.router.srv.mws...)(handler)
}

func (c *Context) ShouldBind(v interface{}) error {
	return c.router.srv.codecs.bodyDecoder(c.req, v)
}

func (c *Context) ShouldBindURI(v interface{}) error {
	return c.router.srv.codecs.varsDecoder(c.req, v)
}

func (c *Context) ShouldBindQuery(v interface{}) error {
	return c.router.srv.codecs.queryDecoder(c.req, v)
}

func (c *Context) Result(v interface{}) error {
	return c.router.srv.codecs.rspEncoder(c.req, c.w, v)
}
