package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/breaker"
	"github.com/go-slark/slark/middleware/logging"
	"github.com/go-slark/slark/middleware/metrics"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/middleware/shedding"
	"github.com/go-slark/slark/middleware/tracing"
	"github.com/go-slark/slark/middleware/validate"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/endpoint"
	"github.com/go-slark/slark/pkg/opentelemetry/metric"
	"github.com/go-slark/slark/transport"
	"github.com/go-slark/slark/transport/http/handler"
	"go.opentelemetry.io/otel/trace"
	"net"
	"net/http"
	"net/url"
)

type Server struct {
	*http.Server
	listener net.Listener
	tls      *tls.Config
	handlers []handler.Middleware
	mws      []middleware.Middleware
	err      error
	network  string
	address  string
	basePath string
	enable   int64
	engine   *gin.Engine
	logger   logger.Logger
	codecs   *Codecs
	headers  []string
}

type ServerOption func(server *Server)

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

func TLS(tls *tls.Config) ServerOption {
	return func(s *Server) {
		s.tls = tls
	}
}

func Handler(handler http.Handler) ServerOption {
	return func(server *Server) {
		server.Handler = handler
	}
}

func Handlers(handlers ...handler.Middleware) ServerOption {
	return func(server *Server) {
		server.handlers = handlers
	}
}

func Middlewares(mws ...middleware.Middleware) ServerOption {
	return func(server *Server) {
		server.mws = mws
	}
}

func Logger(l logger.Logger) ServerOption {
	return func(server *Server) {
		server.logger = l
	}
}

func BasePath(bassPath string) ServerOption {
	return func(server *Server) {
		server.basePath = bassPath
	}
}

func Enable(enable int64) ServerOption {
	return func(s *Server) {
		s.enable = enable
	}
}

func ErrorCodec(ec func(*http.Request, http.ResponseWriter, error)) ServerOption {
	return func(server *Server) {
		server.codecs.errorEncoder = ec
	}
}

func RspCodec(rc func(*http.Request, http.ResponseWriter, interface{}) error) ServerOption {
	return func(server *Server) {
		server.codecs.rspEncoder = rc
	}
}

func Headers(headers []string) ServerOption {
	return func(server *Server) {
		server.headers = headers
	}
}

func NewServer(opts ...ServerOption) *Server {
	engine := gin.New()
	srv := &Server{
		network:  "tcp",
		address:  "0.0.0.0:8080",
		basePath: "/",
		logger:   logger.GetLogger(),
		Server:   &http.Server{},
		handlers: []handler.Middleware{handler.CORS()},
		engine:   engine,
		codecs: &Codecs{
			bodyDecoder:  RequestBodyDecoder,
			varsDecoder:  RequestVarsDecoder,
			queryDecoder: RequestQueryDecoder,
			rspEncoder:   ResponseEncoder,
			errorEncoder: ErrorEncoder,
		},
		headers: []string{utils.Token, utils.Authorization, utils.UserAgent, utils.XForwardedMethod, utils.XForwardedIP, utils.XForwardedURI, utils.Extension},
		mws:     []middleware.Middleware{},
		enable:  0x63, // low -> high
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
	srv.TLSConfig = srv.tls
	srv.handlers = append(srv.handlers, func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trans := &Transport{
				Operation: fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				Req:       Carrier(r.Header),
				Rsp:       Carrier{},
			}
			r = r.WithContext(transport.NewServerContext(r.Context(), trans))
			handler.ServeHTTP(w, r)
		})
	})
	srv.Handler = handler.ComposeMiddleware(srv.engine, srv.handlers...)
	srv.err = srv.listen()
	return srv
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
		Scheme: endpoint.Scheme("http", s.tls == nil),
		Host:   host,
	}
	return u, nil
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Start() error {
	if s.err != nil {
		return s.err
	}
	var err error
	if s.tls != nil {
		err = s.ServeTLS(s.listener, "", "")
	} else {
		err = s.Serve(s.listener)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}
