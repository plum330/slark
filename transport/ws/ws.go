package ws

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/middleware/trace"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
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
	before   func(w http.ResponseWriter, r *http.Request) (interface{}, error)
	after    func(s *Session) error
	opt      *SessionOption
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
		opt: &SessionOption{
			ID:         &gid{},
			in:         1024,
			out:        1024,
			rBuffer:    0,
			wBuffer:    4096,
			hbInterval: 20 * time.Second,
			wTime:      10 * time.Second,
			hsTime:     3 * time.Second,
			rLimit:     51200,
		},
		pool: &sync.Pool{
			New: func() interface{} {
				return new(Session)
			},
		},
		network: "tcp",
		address: "0.0.0.0:0",
		logger:  logger.GetLogger(),
		after: func(s *Session) error {
			s.Close()
			return nil
		},
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

func (s *Server) Handler(hf func(s *Session)) {
	s.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mp := map[string]interface{}{
			"header": r.Header,
			"params": r.URL.Query(),
		}
		var (
			err    error
			result interface{}
		)
		ctx := r.Context()
		if s.before != nil {
			result, err = s.before(w, r)
			ctx = r.Context()
			if err != nil {
				s.logger.Log(ctx, logger.ErrorLevel, mp, "ws establish suspend...")
				return
			}
		}
		s.logger.Log(ctx, logger.DebugLevel, mp, "ws start to establish...")
		session, err := s.NewSession(w, r)
		if err != nil {
			s.logger.Log(ctx, logger.ErrorLevel, mp, "ws establish session error")
			return
		}
		s.logger.Log(ctx, logger.DebugLevel, mp, "ws establish success")
		session.SetContext(ctx)
		if result != nil {
			session.SetCtx(result)
		}
		go func(sess *Session) {
			hf(sess)
			sess.ch <- struct{}{}
		}(session)
		if s.after != nil {
			<-session.ch
			err = s.after(session)
			if err != nil {
				return
			}
		}
	})
}

func (s *Server) Start() error {
	if s.err != nil {
		return s.err
	}
	s.handlers = append(s.handlers, trace.BuildRequestID(), middleware.WrapMiddleware(recovery.Recovery(s.logger)))
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

type SessionOption struct {
	ID
	in         int
	out        int
	rBuffer    int
	wBuffer    int
	hbInterval time.Duration
	wTime      time.Duration
	hsTime     time.Duration
	rLimit     int64
}

type Msg struct {
	Type    int
	Payload []byte
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
	id      string
	context context.Context
	ctx     interface{}
	conn    *websocket.Conn
	in      chan *Msg
	out     chan *Msg
	ch      chan struct{}
	closed  atomic.Bool
	logger  logger.Logger
	pool    *sync.Pool
	l       sync.Mutex
	opt     *SessionOption
	hbTime  int64
	outErr  chan error
	inErr   chan error
}

func (s *Session) reset(conn *websocket.Conn, srv *Server) {
	s.id = srv.opt.NewID()
	s.context = context.Background()
	s.ctx = nil
	s.conn = conn
	s.in = make(chan *Msg, srv.opt.in)
	s.out = make(chan *Msg, srv.opt.out)
	s.ch = make(chan struct{}, 1)
	s.closed.Store(false)
	s.l = sync.Mutex{}
	s.logger = srv.logger
	s.pool = srv.pool
	s.opt = srv.opt
	s.hbTime = time.Now().Unix()
	s.inErr = make(chan error, 1)
	s.outErr = make(chan error, 1)
}

func (s *Session) PingHandler() error {
	if s.closed.Load() {
		return nil
	}
	s.l.Lock()
	defer s.l.Unlock()
	err := s.conn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
	if err != nil {
		return err
	}
	return s.conn.WriteMessage(websocket.PongMessage, nil)
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
	return sess, nil
}

func (s *Session) read() {
	defer func() {
		if e := recover(); e != nil {
			s.logger.Log(s.context, logger.ErrorLevel, map[string]interface{}{"error": e}, "ws read exception")
		}
		s.Close()
	}()
	s.SetHandler()
	s.conn.SetReadLimit(s.opt.rLimit)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
	for {
		msgType, payload, err := s.conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.inErr <- nil
			} else {
				s.inErr <- err
			}
			return
		}
		if s.closed.Load() {
			return
		}
		m := &Msg{
			Type:    msgType,
			Payload: payload,
		}
		s.in <- m
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
	}
}

func (s *Session) write() {
	tk := time.NewTicker(s.opt.hbInterval * 4 / 5)
	defer func() {
		if e := recover(); e != nil {
			s.logger.Log(s.context, logger.ErrorLevel, map[string]interface{}{"error": e}, "ws write exception")
		}
		tk.Stop()
		s.Close()
	}()

	var (
		payload []byte
		msgType int
	)
	for {
		if s.closed.Load() {
			return
		}
		select {
		case m := <-s.out:
			msgType = m.Type
			payload = m.Payload
		case <-tk.C:
			msgType = websocket.PingMessage
			payload = nil
		}
		s.l.Lock()
		_ = s.conn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
		err := s.conn.WriteMessage(msgType, payload)
		s.l.Unlock()
		if err != nil {
			s.outErr <- err
			return
		}
	}
}

func (s *Session) SetHandler() {
	// SetXXHandler work base on ReadMessage()
	s.conn.SetPongHandler(func(msg string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return nil
	})
	s.conn.SetPingHandler(func(msg string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return s.PingHandler()
	})
	s.conn.SetCloseHandler(func(code int, text string) error {
		return nil
	})
}

func (s *Session) Receive() (*Msg, error) {
	select {
	case m := <-s.in:
		return m, nil
	case err := <-s.inErr:
		return nil, err
	}
}

func (s *Session) Send(m *Msg) error {
	var err error
	select {
	case s.out <- m:
	case err = <-s.outErr:
	}
	return err
}

func (s *Session) Close() {
	defer s.pool.Put(s)
	if s.closed.Load() {
		return
	}
	s.closed.Store(true)
	s.l.Lock()
	_ = s.conn.WriteMessage(websocket.CloseMessage, nil)
	s.l.Unlock()
	_ = s.conn.Close()
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

type ID interface {
	NewID() string
}

type XID struct{}

func (x *XID) NewID() string {
	return xid.New().String()
}

type gid struct {
	id uint64
}

func (g *gid) NewID() string {
	id := atomic.AddUint64(&g.id, 1)
	return strconv.FormatUint(id, 36)
}

const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)
