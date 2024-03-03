package grpc

import (
	"github.com/go-slark/slark/transport"
	"google.golang.org/grpc/metadata"
)

type Carrier metadata.MD

func (c Carrier) Set(k string, v string) {
	metadata.MD(c).Set(k, v)
}

func (c Carrier) Add(k string, v string) {
	metadata.MD(c).Append(k, v)
}

func (c Carrier) Get(k string) string {
	v := metadata.MD(c).Get(k)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

type Transport struct {
	operation string
	req       Carrier
	rsp       Carrier
}

func (t *Transport) Kind() string {
	return transport.GRPC
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
