package ws

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/middleware/recovery"
	"github.com/go-slark/slark/middleware/trace"
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
	before   func(w http.ResponseWriter, r *http.Request) (interface{}, error)
	after    func(s *Session) error
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
		s.logger.Log(ctx, logger.InfoLevel, mp, "ws start to establish...")
		session, err := s.NewSession(w, r)
		if err != nil {
			s.logger.Log(ctx, logger.ErrorLevel, mp, "ws establish session error")
			return
		}
		s.logger.Log(ctx, logger.InfoLevel, mp, "ws establish success")
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
	conn      *websocket.Conn
	in        chan *Msg
	out       chan *Msg
	ch        chan struct{}
	closing   chan struct{}
	isClosed  bool
	source    string
	closeTime time.Duration
	logger    logger.Logger
	pool      *sync.Pool
	l         sync.Mutex // avoid close chan duplicated
	opt       *ConnOption
	hbTime    int64
	outErr    chan error
	inErr     chan error
}

func (s *Session) reset(conn *websocket.Conn, srv *Server) {
	s.id = newID()
	s.context = context.Background()
	s.ctx = nil
	s.conn = conn
	s.in = make(chan *Msg, srv.opt.in)
	s.out = make(chan *Msg, srv.opt.out)
	s.ch = make(chan struct{}, 1)
	s.closing = make(chan struct{}, 1)
	s.isClosed = false
	s.source = ""
	s.closeTime = srv.opt.closeTime
	s.logger = srv.logger
	s.pool = srv.pool
	s.opt = srv.opt
	s.hbTime = time.Now().Unix()
	s.inErr = make(chan error, 1)
	s.outErr = make(chan error, 1)
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
		s.conn.SetReadLimit(s.opt.rLimit)
	}
	_ = s.conn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
	for {
		msgType, payload, err := s.conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				s.inErr <- errors.New(504, "read message timeout", "READ_MESSAGE_TIMEOUT").WithError(err)
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.inErr <- errors.InternalServer("read message unexpected close", "READ_MESSAGE_UNEXPECTED_CLOSE").WithError(err)
			} else {
				s.inErr <- errors.New(100, "read message normal close", "READ_MESSAGE_NORMAL_CLOSE").WithError(err)
			}
			s.Close("read_close")
			break
		}
		m := &Msg{
			Type:    msgType,
			Payload: payload,
		}
		select {
		case s.in <- m:
			atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		case <-s.closing:
			s.inErr <- errors.InternalServer(s.source, s.source)
			return
		}
	}
}

func (s *Session) write() {
	tk := time.NewTicker(s.opt.hbInterval * 4 / 5)
	defer func() {
		tk.Stop()
		s.Close("write_close")
	}()

	for {
		select {
		case m := <-s.out:
			_ = s.conn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
			err := s.conn.WriteMessage(m.Type, m.Payload)
			if err != nil {
				s.outErr <- errors.InternalServer("write message exception", "WRITE_MESSAGE_EXCEPTION").WithError(err)
				return
			}
		case <-s.closing:
			s.outErr <- errors.InternalServer(s.source, s.source)
			return
		case <-tk.C:
			_ = s.conn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
			err := s.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				s.outErr <- errors.InternalServer("write ping message exception", "WRITE_PING_MESSAGE_EXCEPTION")
				return
			}
		}
	}
}

func (s *Session) handleHB() {
	// SetXXHandler work base on ReadMessage()
	s.conn.SetPongHandler(func(msg string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.opt.hbInterval))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		return nil
	})
	s.conn.SetPingHandler(func(msg string) error {
		_ = s.conn.SetWriteDeadline(time.Now().Add(s.opt.wTime))
		atomic.StoreInt64(&s.hbTime, time.Now().Unix())
		_ = s.conn.WriteControl(websocket.PongMessage, []byte(msg), time.Now().Add(time.Second))
		return nil
	})
	s.conn.SetCloseHandler(func(code int, text string) error {
		return nil
	})

	for {
		select {
		case <-s.closing:
			err := errors.InternalServer(s.source, s.source)
			s.outErr <- err
			s.inErr <- err
			return

		default:
			ts := atomic.LoadInt64(&s.hbTime)
			if time.Now().Unix()-ts > int64(s.opt.hbInterval.Seconds()) {
				err := errors.InternalServer("session hb exception", "SESSION_HB_EXCEPTION")
				s.outErr <- err
				s.inErr <- err
				s.Close("hb_close")
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
		return nil, <-s.inErr
	}
}

func (s *Session) Send(m *Msg) error {
	var err error
	select {
	case s.out <- m:
	case <-s.closing:
		err = <-s.outErr
	}
	return err
}

func (s *Session) Close(source ...string) {
	defer s.pool.Put(s)
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closed"), time.Now().Add(time.Second))
	time.Sleep(s.closeTime)
	_ = s.conn.Close()
	s.l.Lock()
	defer s.l.Unlock()
	if s.isClosed {
		return
	}
	if len(source) != 0 {
		s.source = source[0]
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
