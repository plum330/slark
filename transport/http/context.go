package http

import (
	"net/http"
)

type Context struct {
	router *Router
	req    *http.Request
	rsp    http.ResponseWriter
}

func (c *Context) Set(req *http.Request, rsp http.ResponseWriter) {
	c.req = req
	c.rsp = rsp
}

func (c *Context) BindBody(v interface{}) error {
	return c.router.srv.Codecs.bodyDecoder(c.req, v)
}

func (c *Context) BindVars(v interface{}) error {
	return c.router.srv.Codecs.varsDecoder(c.req, v)
}

func (c *Context) BindQuery(v interface{}) error {
	return c.router.srv.Codecs.queryDecoder(c.req, v)
}

func (c *Context) Result(code int, v interface{}) error {
	c.rsp.WriteHeader(code)
	return c.router.srv.Codecs.rspEncoder(c.req, c.rsp, v)
}
