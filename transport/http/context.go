package http

import (
	"context"
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
		c.ctx = nil
	} else {
		c.ctx = context.WithValue(c.req.Context(), utils.Token, c.req.Header.Get(utils.Token))
	}
}

func (c *Context) Context() context.Context {
	return c.ctx
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
