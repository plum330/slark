package http

import (
	"github.com/go-slark/slark/transport"
	"net/http"
)

type Carrier http.Header

func (c Carrier) Set(k string, v string) {
	http.Header(c).Set(k, v)
}

func (c Carrier) Add(k string, v string) {
	http.Header(c).Add(k, v)
}

func (c Carrier) Get(k string) string {
	return http.Header(c).Get(k)
}

type Transport struct {
	operation string
	req       Carrier
	rsp       Carrier
	r         *http.Request
	w         http.ResponseWriter
}

func (t *Transport) Kind() string {
	return transport.HTTP
}

func (t *Transport) Operate() string {
	return t.operation
}

func (t *Transport) ReqCarrier() transport.Carrier {
	return t.req
}

func (t *Transport) RspCarrier() transport.Carrier {
	return t.rsp
}
