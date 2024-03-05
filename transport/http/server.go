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
	"github.com/go-slark/slark/middleware/validate"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/transport"
	"github.com/go-slark/slark/transport/http/handler"
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
	maxConn  int
	builtin  int64
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

func MaxConn(size int) ServerOption {
	return func(s *Server) {
		s.maxConn = size
	}
}

func Builtin(builtin int64) ServerOption {
	return func(s *Server) {
		s.builtin = builtin
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
		handlers: []handler.Middleware{},
		engine:   engine,
		codecs: &Codecs{
			bodyDecoder:  RequestBodyDecoder,
			varsDecoder:  RequestVarsDecoder,
			queryDecoder: RequestQueryDecoder,
			rspEncoder:   ResponseEncoder,
			errorEncoder: ErrorEncoder,
		},
		headers: []string{utils.Token, utils.Authorization, utils.UserAgent, utils.XForwardedMethod, utils.XForwardedIP, utils.XForwardedURI, utils.Extension},
		mws:     []middleware.Middleware{validate.Validate()},
		builtin: 0x63, // low -> high
	}
	srv.handlers = []handler.Middleware{
		handler.Trace(),
		handler.WrapMiddleware(logging.Log(srv.logger)),
		handler.WrapMiddleware(metrics.Metrics()),
		handler.MaxConn(srv.logger, srv.maxConn),
		handler.WrapMiddleware(breaker.Breaker()),
		// shedding
		handler.WrapMiddleware(recovery.Recovery(srv.logger)),
		handler.CORS(),
	}
	srv.engine.Use(srv.handle())
	srv.Handler = srv.engine
	srv.TLSConfig = srv.tls
	for _, o := range opts {
		o(srv)
	}
	srv.handlers = utils.Filter(srv.handlers, srv.builtin)
	srv.Handler = handler.ComposeMiddleware(srv.Handler, srv.handlers...)
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
	host, err := utils.ParseAddr(s.listener, s.address)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme: utils.Scheme("http", s.tls == nil),
		Host:   host,
	}
	return u, nil
}

func (s *Server) handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		trans := &Transport{
			operation: fmt.Sprintf("%s %s", req.Method, req.URL.Path),
			req:       Carrier(req.Header),
			rsp:       Carrier{},
		}
		c.Request = req.WithContext(transport.NewServerContext(req.Context(), trans))
	}
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
