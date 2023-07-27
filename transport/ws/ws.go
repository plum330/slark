package ws

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	*http.Server
	handlers []middleware.HTTPMiddleware
	*ConnOption

	ug       *websocket.Upgrader
	listener net.Listener
	handler  http.Handler
	logger   logger.Logger

	network string
	address string
	path    string
	err     error

	// session mng
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		Server: &http.Server{},
		ConnOption: &ConnOption{
			in:         1000,
			out:        1000,
			rBuffer:    1024,
			wBuffer:    1024,
			hbInterval: 15 * time.Second,
			hbTime:     time.Now().Unix(),
			wTime:      10 * time.Second,
			hsTime:     3 * time.Second,
		},
		network: "tcp",
		address: "0.0.0.0:0",
		logger:  logger.GetLogger(),
	}

	for _, opt := range opts {
		opt(srv)
	}

	srv.ug = &websocket.Upgrader{
		HandshakeTimeout: srv.hsTime,
		ReadBufferSize:   srv.rBuffer,
		WriteBufferSize:  srv.wBuffer,
		CheckOrigin: func(r *http.Request) bool {
			// 校验规则
			if r.Method != http.MethodGet {
				return false
			}
			// 允许跨域
			return true
		},
		EnableCompression: false,
	}

	srv.err = srv.listen()
	return srv
}

func (s *Server) Handler(handler http.HandlerFunc) {
	s.handler = handler
}

func (s *Server) Start() error {
	if s.err != nil {
		return s.err
	}

	http.Handle(s.path, middleware.ComposeHTTPMiddleware(s.handler, s.handlers...))
	err := s.Serve(s.listener)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func (s *Server) listen() error {
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}

type ConnOption struct {
	in         int
	out        int
	rBuffer    int
	wBuffer    int
	hbInterval time.Duration
	hbTime     int64
	wTime      time.Duration
	hsTime     time.Duration
	rLimit     int64
}

type Msg struct {
	Type    int
	Payload []byte
	ctx     context.Context
}

type Conn interface {
	ID() string
	Context() interface{}
	SetContext(ctx interface{})
	Close()
	Receive() (*Msg, error)
	Send(m *Msg) error
}

type Session struct {
	id         string
	context    interface{}
	wsConn     *websocket.Conn
	in         chan *Msg
	out        chan *Msg
	closing    chan struct{}
	isClosed   bool
	logger     logger.Logger
	sync.Mutex // avoid close chan duplicated
	*ConnOption
}

func (s *Server) NewSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	ws, err := s.ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	sess := &Session{
		id:         newID(),
		wsConn:     ws,
		in:         make(chan *Msg, s.in),
		out:        make(chan *Msg, s.out),
		closing:    make(chan struct{}, 1),
		logger:     s.logger,
		ConnOption: s.ConnOption,
	}
	go sess.read()
	go sess.write()
	go sess.handleHB()
	return sess, nil
}

func (s *Session) read() {
	if s.rLimit > 0 {
		s.wsConn.SetReadLimit(s.rLimit)
	}
	_ = s.wsConn.SetReadDeadline(time.Now().Add(s.hbInterval))
	for {
		msgType, payload, err := s.wsConn.ReadMessage()
		if err != nil {
			s.Close()
			fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.context), "error": err, "id": s.id}
			s.logger.Log(context.Background(), logger.ErrorLevel, fields, "read message exception")
			break
		}
		m := &Msg{
			Type:    msgType,
			Payload: payload,
			ctx:     context.WithValue(context.Background(), utils.RayID, utils.BuildRequestID()),
		}
		select {
		case s.in <- m:
			atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		case <-s.closing:
			return
		}
	}
}

func (s *Session) write() {
	tk := time.NewTicker(s.hbInterval * 4 / 5)
	defer func() {
		tk.Stop()
		s.Close()
	}()

	for {
		select {
		case m := <-s.out:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(s.wTime))
			err := s.wsConn.WriteMessage(m.Type, m.Payload)
			if err != nil {
				// TODO
				//return
			}
		case <-s.closing:
			return
		case <-tk.C:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(s.wTime))
			err := s.wsConn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				// TODO
				//return
			}
		}
	}
}

func (s *Session) handleHB() {
	s.wsConn.SetPongHandler(func(appData string) error {
		_ = s.wsConn.SetReadDeadline(time.Now().Add(s.hbInterval))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return nil
	})

	for {
		select {
		case <-s.closing:
			return

		default:
			ts := atomic.LoadInt64(&s.hbTime)
			if time.Now().Unix()-ts > int64(s.hbInterval) {
				s.Close()
				return
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *Session) Receive() (*Msg, error) {
	select {
	case m := <-s.in:
		return m, nil
	case <-s.closing:
		return nil, errors.New("conn is closing")
	}
}

func (s *Session) Send(m *Msg) error {
	var err error
	select {
	case s.out <- m:
	case <-s.closing:
		err = errors.New("conn is closing")
	}
	return err
}

func (s *Session) Close() {
	_ = s.wsConn.Close()
	s.Lock()
	defer s.Unlock()
	if s.isClosed {
		return
	}
	close(s.closing)
	s.isClosed = true
}

func (s *Session) SetContext(ctx interface{}) {
	s.context = ctx
}

func (s *Session) Context() interface{} {
	return s.context
}

func (s *Session) ID() string {
	return s.id
}

var connID uint64

func newID() string {
	id := atomic.AddUint64(&connID, 1)
	return strconv.FormatUint(id, 36)
}
