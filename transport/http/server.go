package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/logging"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/middleware/validate"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/transport/http/handler"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
	listener net.Listener
	handlers []handler.Middleware
	mws      []middleware.Middleware
	err      error
	network  string
	address  string
	basePath string
	Engine   *gin.Engine
	logger   logger.Logger
	Codecs   *Codecs
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

func ErrorCodec(ec func(*http.Request, http.ResponseWriter, error)) ServerOption {
	return func(server *Server) {
		server.Codecs.errorEncoder = ec
	}
}

func RspCodec(rc func(*http.Request, http.ResponseWriter, interface{}) error) ServerOption {
	return func(server *Server) {
		server.Codecs.rspEncoder = rc
	}
}

func Headers(headers []string) ServerOption {
	return func(server *Server) {
		server.headers = headers
	}
}

// trace -> log -> metric -> breaker -> recovery -> ...

func NewServer(opts ...ServerOption) *Server {
	engine := gin.New()
	srv := &Server{
		network:  "tcp",
		address:  "0.0.0.0:0",
		basePath: "/",
		logger:   logger.GetLogger(),
		Server:   &http.Server{},
		handlers: []handler.Middleware{handler.BuildRequestID(), handler.CORS()},
		Engine:   engine,
		Codecs: &Codecs{
			bodyDecoder:  RequestBodyDecoder,
			varsDecoder:  RequestVarsDecoder,
			queryDecoder: RequestQueryDecoder,
			rspEncoder:   ResponseEncoder,
			errorEncoder: ErrorEncoder,
		},
		headers: []string{utils.Token, utils.Authorization, utils.UserAgent, utils.XForwardedMethod, utils.XForwardedIP, utils.XForwardedURI, utils.Extension},
	}
	srv.Handler = srv.Engine
	for _, o := range opts {
		o(srv)
	}
	srv.mws = append(srv.mws, logging.Log(srv.logger), validate.Validate(), recovery.Recovery(srv.logger))
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

func (s *Server) Start() error {
	if s.err != nil {
		return s.err
	}
	err := s.Serve(s.listener)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}
