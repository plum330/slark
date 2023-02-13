package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/transport/http/handler"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
	listener net.Listener
	handlers []filter.Handler // cors...
	err      error
	network  string
	address  string
	Engine   *gin.Engine
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

func Handlers(handlers ...filter.Handler) ServerOption {
	return func(server *Server) {
		server.handlers = handlers
	}
}

func NewServer(opts ...ServerOption) *Server {
	engine := gin.New()
	srv := &Server{
		network: "tcp",
		address: "0.0.0.0:0",
		Server:  &http.Server{},
		Engine:  engine,
	}
	srv.Handler = srv.Engine
	for _, o := range opts {
		o(srv)
	}
	srv.Handler = filter.Handle(srv.Handler, srv.handlers...)
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
