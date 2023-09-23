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
	opt      *ConnOption
	ug       *websocket.Upgrader
	listener net.Listener
	handler  http.Handler
	logger   logger.Logger
	pool     *sync.Pool
	network  string
	address  string
	path     string
	err      error
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		Server: &http.Server{},
		opt: &ConnOption{
			in:         1000,
			out:        1000,
			rBuffer:    1024,
			wBuffer:    1024,
			hbInterval: 60 * time.Second,
			wTime:      10 * time.Second,
			hsTime:     3 * time.Second,
			closeTime:  500 * time.Millisecond,
		},
		pool: &sync.Pool{
			New: func() interface{} {
				return new(Session)
			},
		},
		network: "tcp",
		address: "0.0.0.0:0",
		logger:  logger.GetLogger(),
	}

	for _, opt := range opts {
		opt(srv)
	}

	srv.ug = &websocket.Upgrader{
		HandshakeTimeout: srv.opt.hsTime,
		ReadBufferSize:   srv.opt.rBuffer,
		WriteBufferSize:  srv.opt.wBuffer,
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
	wTime      time.Duration
	hsTime     time.Duration
	closeTime  time.Duration
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
	id        string
	context   context.Context
	ctx       interface{}
	wsConn    *websocket.Conn
	in        chan *Msg
	out       chan *Msg
	closing   chan struct{}
	isClosed  bool
	closeTime time.Duration
	logger    logger.Logger
	pool      *sync.Pool
	l         sync.Mutex // avoid close chan duplicated
	opt       *ConnOption
	hbTime    int64
}

func (s *Session) reset(conn *websocket.Conn, srv *Server) {
	s.id = newID()
	s.context = context.Background()
	s.wsConn = conn
	s.in = make(chan *Msg, srv.opt.in)
	s.out = make(chan *Msg, srv.opt.out)
	s.closing = make(chan struct{}, 1)
	s.closeTime = srv.opt.closeTime
	s.logger = srv.logger
	s.pool = srv.pool
	s.opt = srv.opt
	s.hbTime = time.Now().Unix()
}

func (s *Server) NewSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	ws, err := s.ug.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	sess := s.pool.Get().(*Session)
	sess.reset(ws, s)
	go sess.read()
	go sess.write()
	go sess.handleHB()
	return sess, nil
}

func (s *Session) read() {
	if s.opt.rLimit > 0 {
		s.wsConn.SetReadLimit(s.opt.rLimit)
	}
	_ = s.wsConn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
	for {
		msgType, payload, err := s.wsConn.ReadMessage()
		if err != nil {
			fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "error": fmt.Sprintf("%+v", err), "id": s.id}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				s.logger.Log(s.context, logger.ErrorLevel, fields, "read message timeout")
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.logger.Log(s.context, logger.ErrorLevel, fields, "read message unexpected close")
			} else {
				s.logger.Log(s.context, logger.DebugLevel, fields, "read message close")
			}
			s.Close()
			break
		}
		m := &Msg{
			Type:    msgType,
			Payload: payload,
			ctx:     context.WithValue(context.WithValue(context.Background(), utils.RayID, utils.BuildRequestID()), utils.Token, s.context.Value(utils.Token)),
		}
		select {
		case s.in <- m:
			atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		case <-s.closing:
			fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "id": s.id}
			s.logger.Log(s.context, logger.WarnLevel, fields, "session read closing")
			return
		}
	}
}

func (s *Session) write() {
	tk := time.NewTicker(s.opt.hbInterval * 4 / 5)
	defer func() {
		tk.Stop()
		s.Close()
	}()

	for {
		select {
		case m := <-s.out:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
			err := s.wsConn.WriteMessage(m.Type, m.Payload)
			if err != nil {
				fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "error": fmt.Sprintf("%+v", err), "id": s.id}
				s.logger.Log(s.context, logger.ErrorLevel, fields, "write message exception")
				return
			}
		case <-s.closing:
			fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "id": s.id}
			s.logger.Log(s.context, logger.WarnLevel, fields, "session write closing")
			return
		case <-tk.C:
			_ = s.wsConn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
			err := s.wsConn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "error": fmt.Sprintf("%+v", err), "id": s.id}
				s.logger.Log(s.context, logger.ErrorLevel, fields, "write ping message exception")
				return
			}
		}
	}
}

func (s *Session) handleHB() {
	s.wsConn.SetPongHandler(func(appData string) error {
		_ = s.wsConn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return nil
	})
	s.wsConn.SetPingHandler(func(appData string) error {
		_ = s.wsConn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return nil
	})
	s.wsConn.SetCloseHandler(func(code int, text string) error {
		return nil
	})

	for {
		select {
		case <-s.closing:
			fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "id": s.id}
			s.logger.Log(s.context, logger.WarnLevel, fields, "session hb closing")
			return

		default:
			ts := atomic.LoadInt64(&s.hbTime)
			if time.Now().Unix()-ts > int64(s.opt.hbInterval.Seconds()) {
				s.Close()
				fields := map[string]interface{}{"context": fmt.Sprintf("%+v", s.ctx), "id": s.id}
				s.logger.Log(s.context, logger.WarnLevel, fields, "session hb exception")
				return
			}
			time.Sleep(s.opt.hbInterval * 1 / 10) // 10%
		}
	}
}

func (s *Session) Receive() (*Msg, error) {
	select {
	case m := <-s.in:
		return m, nil
	case <-s.closing:
		return nil, errors.New("conn receive is closing")
	}
}

func (s *Session) Send(m *Msg) error {
	var err error
	select {
	case s.out <- m:
	case <-s.closing:
		err = errors.New("conn send is closing")
	}
	return err
}

func (s *Session) Close() {
	defer s.pool.Put(s)
	_ = s.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closed"))
	time.Sleep(s.closeTime)
	_ = s.wsConn.Close()
	s.l.Lock()
	defer s.l.Unlock()
	if s.isClosed {
		return
	}
	close(s.closing)
	s.isClosed = true
}

func (s *Session) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *Session) Context() context.Context {
	return s.context
}

func (s *Session) SetCtx(ctx interface{}) {
	s.ctx = ctx
}

func (s *Session) Ctx() interface{} {
	return s.ctx
}

func (s *Session) ID() string {
	return s.id
}

var connID uint64

func newID() string {
	id := atomic.AddUint64(&connID, 1)
	return strconv.FormatUint(id, 36)
}
