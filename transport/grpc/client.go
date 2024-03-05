package grpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/breaker"
	"github.com/go-slark/slark/middleware/logging"
	"github.com/go-slark/slark/middleware/metrics"
	"github.com/go-slark/slark/middleware/tracing"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/registry"
	"github.com/go-slark/slark/transport/grpc/balancer/node"
	"github.com/go-slark/slark/transport/grpc/resolver"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

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
	keepalive Keepalive
	strategy  []Strategy
	addr      string
	size      int // subset size
	subset    resolver.Subset
	insecure  bool
	builtin   int64
	tm        time.Duration
	logger    logger.Logger
	tls       *tls.Config
	opts      []grpc.DialOption
	unary     []grpc.UnaryClientInterceptor
	stream    []grpc.StreamClientInterceptor
	mw        []middleware.Middleware
	discovery registry.Discovery
	filters   []node.Filter
}

type Option func(*option)

func WithTimeout(tm time.Duration) Option {
	return func(o *option) {
		o.tm = tm
	}
}

func WithLogger(l logger.Logger) Option {
	return func(o *option) {
		o.logger = l
	}
}

func WithAddr(addr string) Option {
	return func(o *option) {
		o.addr = addr
	}
}

func WithBuiltin(builtin int64) Option {
	return func(o *option) {
		o.builtin = builtin
	}
}

func WithSize(size int) Option {
	return func(o *option) {
		o.size = size
	}
}

func WithSubset(subset resolver.Subset) Option {
	return func(o *option) {
		o.subset = subset
	}
}

func WithInsecure(insecure bool) Option {
	return func(o *option) {
		o.insecure = insecure
	}
}

func WithTLS(tls *tls.Config) Option {
	return func(o *option) {
		o.tls = tls
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

func WithFilters(filters []node.Filter) Option {
	return func(o *option) {
		o.filters = filters
	}
}

func WithKeepalive(keepalive Keepalive) Option {
	return func(o *option) {
		o.keepalive = keepalive
	}
}

/*
{
	"methodConfig": [{
		  "retryPolicy": {
			  "MaxAttempts": 3,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
	}}]
}
MaxAttempts一次原始请求，2次重试，最大值5
InitialBakckoff, BackoffMultiplier, MaxBackoff计算重试间隔:第一次重试间隔random(0, InitialBakckoff), 第n次重试间隔random(0, min( InitialBakckoff*BackoffMultiplier*(n-1) , MaxBackoff))
*/

func WithStrategy(strategy []Strategy) Option {
	return func(o *option) {
		o.strategy = strategy
	}
}

func Dial(ctx context.Context, opts ...Option) (*grpc.ClientConn, error) {
	opt := &option{
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
		size:     32,
		subset:   &resolver.Shuffle{},
		builtin:  0x03,
	}
	opt.mw = []middleware.Middleware{
		tracing.Trace(trace.SpanKindClient),
		logging.Log(opt.logger),
		metrics.Metrics(),
		breaker.Breaker(),
	}
	for _, o := range opts {
		o(opt)
	}
	opt.mw = utils.Filter(opt.mw, opt.builtin)
	unary := []grpc.UnaryClientInterceptor{unaryClientInterceptor(opt)}
	stream := []grpc.StreamClientInterceptor{streamClientInterceptor(opt)}
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
	if opt.tls != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(opt.tls)))
	}
	if len(opt.opts) > 0 {
		dialOpts = append(dialOpts, opt.opts...)
	}

	if opt.discovery != nil {
		dialOpts = append(dialOpts, grpc.WithResolvers(resolver.NewBuilder(opt.discovery, resolver.WithInsecure(opt.insecure), resolver.WithSize(opt.size), resolver.WithSubSet(opt.subset))))
	}
	return grpc.DialContext(ctx, opt.addr, dialOpts...)
}
