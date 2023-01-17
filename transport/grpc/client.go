package grpc

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ctx struct {
	tm time.Duration
	c  context.Context
	f  context.CancelFunc
}

type Client struct {
	*grpc.ClientConn
	listener net.Listener
	ctx      ctx
	err      error
	address  string
	opts     []grpc.DialOption
	unary    []grpc.UnaryClientInterceptor
	stream   []grpc.StreamClientInterceptor
}

func NewClient(opts ...ClientOption) *Client {
	cli := &Client{
		ctx:     ctx{c: context.TODO(), f: nil, tm: 0},
		address: "0.0.0.0:0",
	}
	for _, o := range opts {
		o(cli)
	}

	if cli.ctx.tm != 0 {
		cli.ctx.c, cli.ctx.f = context.WithTimeout(context.Background(), cli.ctx.tm)
		defer cli.ctx.f()
	}

	var grpcOpts []grpc.DialOption
	if len(cli.unary) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithChainUnaryInterceptor(cli.unary...))
	}
	if len(cli.stream) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithChainStreamInterceptor(cli.stream...))
	}
	if len(cli.opts) > 0 {
		grpcOpts = append(grpcOpts, cli.opts...)
	}

	conn, err := grpc.DialContext(cli.ctx.c, cli.address, grpcOpts...)
	cli.err = err
	cli.ClientConn = conn
	return cli
}

func (c *Client) Stop() error {
	return c.Close()
}

type ClientOption func(*Client)

func ClientOptions(opts []grpc.DialOption) ClientOption {
	return func(client *Client) {
		client.opts = opts
	}
}

func WithAddr(addr string) ClientOption {
	return func(client *Client) {
		client.address = addr
	}
}

func WithTimeout(tm time.Duration) ClientOption {
	return func(client *Client) {
		client.ctx.tm = tm
	}
}

func WithUnaryInterceptor(unary []grpc.UnaryClientInterceptor) ClientOption {
	return func(client *Client) {
		client.unary = unary
	}
}

func WithStreamInterceptor(stream []grpc.StreamClientInterceptor) ClientOption {
	return func(client *Client) {
		client.stream = stream
	}
}
