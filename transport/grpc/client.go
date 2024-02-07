package grpc

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

type ctx struct {
	tm time.Duration
	c  context.Context
	f  context.CancelFunc
}

type Keepalive struct {
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
	KeepaliveStream  bool
}

type Strategy struct {
	Name  string
	Value string
}

type option struct {
	ctx       ctx
	keepalive Keepalive
	strategy  []Strategy
	addr      string
	insecure  bool
	opts      []grpc.DialOption
	unary     []grpc.UnaryClientInterceptor
	stream    []grpc.StreamClientInterceptor
	mw        []middleware.Middleware
	discovery registry.Discovery
}

type Option func(*option)

func WithTimeout(tm time.Duration) Option {
	return func(o *option) {
		o.ctx.tm = tm
	}
}

func WithAddr(addr string) Option {
	return func(o *option) {
		o.addr = addr
	}
}

func WithMiddleware(mw []middleware.Middleware) Option {
	return func(o *option) {
		o.mw = mw
	}
}

func WithUnaryInterceptor(unary []grpc.UnaryClientInterceptor) Option {
	return func(o *option) {
		o.unary = unary
	}
}

func appendUnaryInterceptor(unary []grpc.UnaryClientInterceptor) Option {
	return func(o *option) {
		o.unary = append(o.unary, unary...)
	}
}

func WithStreamInterceptor(stream []grpc.StreamClientInterceptor) Option {
	return func(o *option) {
		o.stream = stream
	}
}

func WithDiscovery(discovery registry.Discovery) Option {
	return func(o *option) {
		o.discovery = discovery
	}
}

func WithKeepalive(keepalive Keepalive) Option {
	return func(o *option) {
		o.keepalive = keepalive
	}
}

func WithStrategy(strategy []Strategy) Option {
	return func(o *option) {
		o.strategy = strategy
	}
}

func Dial(opts ...Option) (*grpc.ClientConn, error) {
	opt := &option{
		ctx: ctx{
			c:  context.TODO(),
			f:  nil,
			tm: 5 * time.Second,
		},
		keepalive: Keepalive{
			KeepaliveTime:    2 * time.Minute,
			KeepaliveTimeout: 2 * time.Second,
			KeepaliveStream:  true,
		},
		addr: "0.0.0.0:0",
		strategy: []Strategy{
			{
				Name:  fmt.Sprintf(`"%s"`, "loadBalancingConfig"),
				Value: `[ {"round_robin": {} } ]`,
			},
		},
		insecure: true,
	}
	for _, o := range opts {
		o(opt)
	}

	if opt.ctx.tm != 0 {
		opt.ctx.c, opt.ctx.f = context.WithTimeout(context.Background(), opt.ctx.tm)
		defer opt.ctx.f()
	}

	unary := []grpc.UnaryClientInterceptor{unaryClientInterceptor(opt.mw...)}
	stream := []grpc.StreamClientInterceptor{streamClientInterceptor(opt.mw)}
	if len(opt.unary) > 0 {
		unary = append(unary, opt.unary...)
	}
	if len(opt.stream) > 0 {
		stream = append(stream, opt.stream...)
	}
	dialOpts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                opt.keepalive.KeepaliveTime,
			Timeout:             opt.keepalive.KeepaliveTimeout,
			PermitWithoutStream: opt.keepalive.KeepaliveStream,
		}),
		grpc.WithChainUnaryInterceptor(unary...),
		grpc.WithChainStreamInterceptor(stream...),
	}

	if opt.insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if len(opt.strategy) > 0 {
		var buf bytes.Buffer
		buf.WriteString("{")
		for index, s := range opt.strategy {
			buf.WriteString(s.Name)
			buf.WriteString(":")
			buf.WriteString(s.Value)
			if index < len(opt.strategy)-1 {
				buf.WriteString(",")
			}
		}
		buf.WriteString("}")
		dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(buf.String()))
	}

	if len(opt.opts) > 0 {
		dialOpts = append(dialOpts, opt.opts...)
	}

	if opt.discovery != nil {
		dialOpts = append(dialOpts, grpc.WithResolvers(NewBuilder(opt.discovery)))
	}
	return grpc.DialContext(opt.ctx.c, opt.addr, dialOpts...)
}
