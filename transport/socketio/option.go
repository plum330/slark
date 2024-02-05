package socketio

import (
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/transport/http/handler"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/session"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/rs/xid"
	"time"
)

type Option func(*Server)

func Path(path string) Option {
	return func(s *Server) {
		s.path = path
	}
}

func Addr(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

func Network(network string) Option {
	return func(s *Server) {
		s.network = network
	}
}

func Logger(l logger.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

func Handlers(handlers []handler.Middleware) Option {
	return func(s *Server) {
		s.handlers = handlers
	}
}

func Options(opts ...EIOOption) Option {
	return func(s *Server) {
		for _, opt := range opts {
			opt(s.eio)
		}
	}
}

type EIOOption func(options *engineio.Options)

type XID struct{}

func (x *XID) NewID() string {
	return xid.New().String()
}

func ID(id session.IDGenerator) EIOOption {
	return func(o *engineio.Options) {
		o.SessionIDGenerator = id
	}
}

func Transports(transports []transport.Transport) EIOOption {
	return func(o *engineio.Options) {
		o.Transports = transports
	}
}

func PingTimeout(tm time.Duration) EIOOption {
	return func(o *engineio.Options) {
		o.PingTimeout = tm
	}
}

func PingInterval(interval time.Duration) EIOOption {
	return func(o *engineio.Options) {
		o.PingInterval = interval
	}
}

func RedisOption(opts ...RedisOptions) Option {
	return func(s *Server) {
		for _, opt := range opts {
			opt(s.adapter)
		}
	}
}

type RedisOptions func(options *socketio.RedisAdapterOptions)

func Address(addr string) RedisOptions {
	return func(o *socketio.RedisAdapterOptions) {
		o.Addr = addr
	}
}

func Prefix(prefix string) RedisOptions {
	return func(o *socketio.RedisAdapterOptions) {
		o.Prefix = prefix
	}
}

func Password(password string) RedisOptions {
	return func(o *socketio.RedisAdapterOptions) {
		o.Password = password
	}
}

func DB(db int) RedisOptions {
	return func(o *socketio.RedisAdapterOptions) {
		o.DB = db
	}
}

func WithNetwork(network string) RedisOptions {
	return func(o *socketio.RedisAdapterOptions) {
		o.Network = network
	}
}

func Adapter(enable bool) Option {
	return func(s *Server) {
		s.enable = enable
	}
}
