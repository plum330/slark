package transport

import (
	"context"
	"net/url"

	_ "github.com/go-slark/slark/encoding/form"
	_ "github.com/go-slark/slark/encoding/json"
	_ "github.com/go-slark/slark/encoding/msgpack"
	_ "github.com/go-slark/slark/encoding/proto"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type Endpoint interface {
	Endpoint() (*url.URL, error)
}

const (
	HTTP = "http"
	GRPC = "grpc"
)

type Carrier interface {
	Set(k string, v string)
	Add(k string, v string)
	Get(k string) string
}

type Transporter interface {
	Kind() string
	Operate() string
	ReqCarrier() Carrier
	RspCarrier() Carrier
}

type clientContextKey struct{}

func NewClientContext(ctx context.Context, trans Transporter) context.Context {
	return context.WithValue(ctx, clientContextKey{}, trans)
}

func FromClientContext(ctx context.Context) (Transporter, bool) {
	trans, ok := ctx.Value(clientContextKey{}).(Transporter)
	return trans, ok
}

type serverContextKey struct{}

func NewServerContext(ctx context.Context, trans Transporter) context.Context {
	return context.WithValue(ctx, serverContextKey{}, trans)
}

func FromServerContext(ctx context.Context) (Transporter, bool) {
	trans, ok := ctx.Value(serverContextKey{}).(Transporter)
	return trans, ok
}
