package socketio

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/transport/http/handler"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"net"
	"net/http"
)

type Server struct {
	*socketio.Server
	listener net.Listener
	handlers []handler.Middleware
	logger   logger.Logger
	eio      *engineio.Options
	adapter  *socketio.RedisAdapterOptions
	enable   bool
	path     string
	address  string
	network  string
	err      error
}

func NewServer(opts ...Option) *Server {
	srv := &Server{
		eio: &engineio.Options{
			Transports: []transport.Transport{
				&websocket.Transport{
					CheckOrigin: func(r *http.Request) bool {
						return true
					},
				},
				&polling.Transport{
					CheckOrigin: func(r *http.Request) bool {
						return true
					},
				},
			},
		},
		adapter: &socketio.RedisAdapterOptions{
			Addr:    "0.0.0:2379",
			Network: "tcp",
			DB:      0,
		},
		logger:  logger.GetLogger(),
		network: "tcp",
		address: "0.0.0.0:0",
		path:    "/socket.io/",
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.Server = socketio.NewServer(srv.eio)
	if srv.enable {
		_, srv.err = srv.Adapter(srv.adapter)
		if srv.err != nil {
			return srv
		}
	}
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
	go func() {
		if err := s.Serve(); err != nil {
			s.logger.Log(context.TODO(), logger.FatalLevel, map[string]interface{}{"error": err}, "socket.io listen error")
			return
		}
	}()
	s.handlers = append(s.handlers, handler.BuildRequestID(), handler.WrapMiddleware(recovery.Recovery(s.logger)))
	http.Handle(s.path, handler.ComposeMiddleware(s, s.handlers...))
	err := http.Serve(s.listener, nil)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(_ context.Context) error {
	return s.Close()
}

// tips: different namespace can represent different biz

type Conn = socketio.Conn

func (s *Server) OnConnect(namespace string, f func(Conn) error) {
	s.Server.OnConnect(namespace, f)
}

func (s *Server) OnDisconnect(namespace string, f func(Conn, string)) {
	s.Server.OnDisconnect(namespace, f)
}

func (s *Server) OnEvent(namespace, event string, f interface{}) {
	s.Server.OnEvent(namespace, event, f)
}

func (s *Server) OnError(namespace string, f func(Conn, error)) {
	s.Server.OnError(namespace, f)
}
