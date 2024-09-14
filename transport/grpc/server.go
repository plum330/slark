package grpc

import (
	"context"
	"crypto/tls"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/flexible/breaker"
	"github.com/go-slark/slark/middleware/flexible/shedding"
	"github.com/go-slark/slark/middleware/logging"
	"github.com/go-slark/slark/middleware/metrics"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/middleware/tracing"
	"github.com/go-slark/slark/middleware/validate"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/endpoint"
	"github.com/go-slark/slark/pkg/opentelemetry/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"net/url"
	"time"
)

type Server struct {
	*grpc.Server
	health   *health.Server
	listener net.Listener
	tls      *tls.Config
	err      error
	logger   logger.Logger
	network  string
	address  string
	enable   int64
	timeout  time.Duration
	mws      []middleware.Middleware
	opts     []grpc.ServerOption
	unary    []grpc.UnaryServerInterceptor
	stream   []grpc.StreamServerInterceptor
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: "0.0.0.0:9090",
		health:  health.NewServer(),
		logger:  logger.GetLogger(),
		opts:    ServerOpts(),
		enable:  0x63,
	}
	srv.mws = []middleware.Middleware{
		tracing.Trace(trace.SpanKindServer),
		logging.Log(middleware.Server, srv.logger),
		metrics.Metrics(middleware.Server, metric.WithHistogram(metric.RequestDurationHistogram())),
		breaker.Breaker(),
		shedding.Limit(),
		recovery.Recovery(srv.logger),
		validate.Validate(),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.mws = utils.Filter(srv.mws, srv.enable)
	var grpcOpts []grpc.ServerOption
	srv.unary = append(srv.unary, srv.unaryServerInterceptor())
	srv.stream = append(srv.stream, srv.streamServerInterceptor())
	if len(srv.unary) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(srv.unary...))
	}
	if len(srv.stream) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainStreamInterceptor(srv.stream...))
	}
	if srv.tls != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tls)))
	}
	if len(srv.opts) > 0 {
		grpcOpts = append(grpcOpts, srv.opts...)
	}

	srv.Server = grpc.NewServer(grpcOpts...)
	srv.err = srv.listen()
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server)
	return srv
}

func (s *Server) Start() error {
	if s.err != nil {
		return s.err
	}
	s.health.Resume()
	return s.Serve(s.listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.health.Shutdown()
	s.GracefulStop()
	return nil
}

func (s *Server) listen() error {
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}

func (s *Server) Endpoint() (*url.URL, error) {
	host, err := endpoint.ParseAddr(s.listener, s.address)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme: endpoint.Scheme("grpc", s.tls == nil),
		Host:   host,
	}
	return u, nil
}

type ServerOption func(*Server)

func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func Timeout(tm time.Duration) ServerOpt {
	return func(s *serverOpt) {
		s.timeout = tm
	}
}

func Listener(l net.Listener) ServerOption {
	return func(s *Server) {
		s.listener = l
	}
}

func TLS(tls *tls.Config) ServerOption {
	return func(s *Server) {
		s.tls = tls
	}
}

func Logger(logger logger.Logger) ServerOption {
	return func(server *Server) {
		server.logger = logger
	}
}

func Enable(enable int64) ServerOption {
	return func(server *Server) {
		server.enable = enable
	}
}

func UnaryInterceptor(u []grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unary = u
	}
}

func StreamInterceptor(s []grpc.StreamServerInterceptor) ServerOption {
	return func(server *Server) {
		server.stream = s
	}
}

func ServerOptions(opts []grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.opts = opts
	}
}

func Middleware(mws []middleware.Middleware) ServerOption {
	return func(server *Server) {
		server.mws = mws
	}
}

type serverOpt struct {
	maxConnectionIdle     time.Duration
	maxConnectionAge      time.Duration
	maxConnectionAgeGrace time.Duration
	time                  time.Duration
	timeout               time.Duration
	minTime               time.Duration
	permitWithoutStream   bool
}

type ServerOpt func(option *serverOpt)

func MaxConnectionIdle(idle time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.maxConnectionIdle = idle
	}
}

func MaxConnectionAge(age time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.maxConnectionAge = age
	}
}

func MaxConnectionAgeGrace(ag time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.maxConnectionAgeGrace = ag
	}
}

func AliveTime(tm time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.time = tm
	}
}

func AliveTimeout(tm time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.timeout = tm
	}
}

func MinTime(mt time.Duration) ServerOpt {
	return func(option *serverOpt) {
		option.minTime = mt
	}
}

func PermitWithoutStream(pws bool) ServerOpt {
	return func(option *serverOpt) {
		option.permitWithoutStream = pws
	}
}

func ServerOpts(opts ...ServerOpt) []grpc.ServerOption {
	o := &serverOpt{
		maxConnectionIdle:     5 * time.Minute,
		maxConnectionAge:      0,
		maxConnectionAgeGrace: 5 * time.Second,
		time:                  2 * time.Minute,
		timeout:               2 * time.Second,
		minTime:               5 * time.Second,
		permitWithoutStream:   true,
	}
	for _, opt := range opts {
		opt(o)
	}
	opt := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     o.maxConnectionIdle,
			MaxConnectionAge:      o.maxConnectionAge,
			MaxConnectionAgeGrace: o.maxConnectionAgeGrace,
			Time:                  o.time,
			Timeout:               o.timeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             o.minTime,
			PermitWithoutStream: o.permitWithoutStream,
		}),
	}
	return opt
}
