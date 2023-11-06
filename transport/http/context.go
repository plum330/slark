package http

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"net/http"
)

type Context struct {
	router *Router
	req    *http.Request
	rsp    http.ResponseWriter
	ctx    context.Context
}

func (c *Context) Set(req *http.Request, rsp http.ResponseWriter) {
	c.req = req
	c.rsp = rsp
	if c.req == nil {
		c.ctx = context.Background()
	} else {
		c.ctx = c.req.Context()
		for _, hk := range c.router.srv.headers {
			hv := c.req.Header.Get(hk)
			if len(hv) == 0 {
				continue
			}
			c.ctx = context.WithValue(c.ctx, hk, hv)
		}
	}
}

func (c *Context) SetMethod(method, path string) {
	c.ctx = context.WithValue(c.ctx, utils.Method, fmt.Sprintf("%s:%s", method, path))
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
	return c.router.srv.Codecs.rspEncoder(c.req, c.rsp, v)
}
