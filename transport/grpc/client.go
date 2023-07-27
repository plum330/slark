package grpc

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/registry"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
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
	listener  net.Listener
	ctx       ctx
	err       error
	address   string
	opts      []grpc.DialOption
	unary     []grpc.UnaryClientInterceptor
	stream    []grpc.StreamClientInterceptor
	mw        []middleware.Middleware
	discovery registry.Discovery
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
	unary := []grpc.UnaryClientInterceptor{cli.unaryClientInterceptor()}
	if len(cli.unary) > 0 {
		unary = append(unary, cli.unary...)
	}
	grpcOpts = append(grpcOpts, grpc.WithChainUnaryInterceptor(unary...))
	if len(cli.stream) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithChainStreamInterceptor(cli.stream...))
	}
	if len(cli.opts) > 0 {
		grpcOpts = append(grpcOpts, cli.opts...)
	}

	if cli.discovery != nil {
		grpcOpts = append(grpcOpts, grpc.WithResolvers(NewBuilder(cli.discovery)))
	}

	conn, err := grpc.DialContext(cli.ctx.c, cli.address, grpcOpts...)
	cli.err = err
	cli.ClientConn = conn
	return cli
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

func WithMiddle(mw []middleware.Middleware) ClientOption {
	return func(client *Client) {
		client.mw = mw
	}
}

func Discovery(discovery registry.Discovery) ClientOption {
	return func(client *Client) {
		client.discovery = discovery
	}
}

type dialOption struct {
	retry            uint
	retryTimeout     time.Duration
	waitBetween      time.Duration
	jitter           float64
	timeout          time.Duration
	keepaliveTime    time.Duration
	keepaliveTimeout time.Duration
	keepaliveStream  bool
}

type DialOpt func(do *dialOption)

func Retry(retry uint) DialOpt {
	return func(do *dialOption) {
		do.retry = retry
	}
}

func RetryTimeout(tm time.Duration) DialOpt {
	return func(do *dialOption) {
		do.retryTimeout = tm
	}
}

func WaitBetween(wb time.Duration) DialOpt {
	return func(do *dialOption) {
		do.waitBetween = wb
	}
}

func Jitter(j float64) DialOpt {
	return func(do *dialOption) {
		do.jitter = j
	}
}

func Timeout(tm time.Duration) DialOpt {
	return func(do *dialOption) {
		do.timeout = tm
	}
}

func KeepaliveTime(tm time.Duration) DialOpt {
	return func(do *dialOption) {
		do.keepaliveTime = tm
	}
}

func KeepaliveTimeout(tm time.Duration) DialOpt {
	return func(do *dialOption) {
		do.keepaliveTimeout = tm
	}
}

func KeepaliveStream(stream bool) DialOpt {
	return func(do *dialOption) {
		do.keepaliveStream = stream
	}
}

func DialOpts(opt ...DialOpt) []grpc.DialOption {
	do := &dialOption{
		retry:            3,
		retryTimeout:     time.Second * 2,
		waitBetween:      time.Second / 2,
		jitter:           0.2,
		timeout:          5 * time.Second,
		keepaliveTime:    10 * time.Second,
		keepaliveTimeout: time.Second,
		keepaliveStream:  true,
	}
	for _, o := range opt {
		o(do)
	}
	retryOps := []grpc_retry.CallOption{
		grpc_retry.WithMax(do.retry),
		grpc_retry.WithPerRetryTimeout(do.retryTimeout),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(do.waitBetween, do.jitter)),
	}
	retry := grpc_retry.UnaryClientInterceptor(retryOps...)
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(retry, UnaryClientTimeout(do.timeout)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                do.keepaliveTime,
			Timeout:             do.keepaliveTimeout,
			PermitWithoutStream: do.keepaliveStream,
		}),
	}
	return opts
}
