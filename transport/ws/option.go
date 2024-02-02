package ws

import (
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"net/http"
	"time"
)

type ServerOption func(*Server)

func Logger(l logger.Logger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

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

func Timeout(rTimeout, wTimeout time.Duration) ServerOption {
	return func(s *Server) {
		s.Server.ReadTimeout = rTimeout
		s.Server.WriteTimeout = wTimeout
	}
}

func Path(path string) ServerOption {
	return func(s *Server) {
		s.path = path
	}
}

func Before(before func(w http.ResponseWriter, r *http.Request) (interface{}, error)) ServerOption {
	return func(s *Server) {
		s.before = before
	}
}

func After(after func(s *Session) error) ServerOption {
	return func(s *Server) {
		s.after = after
	}
}

func ConnOpt(opts ...Option) ServerOption {
	return func(server *Server) {
		for _, opt := range opts {
			opt(server.opt)
		}
	}
}

func Handlers(handlers ...middleware.HTTPMiddleware) ServerOption {
	return func(server *Server) {
		server.handlers = handlers
	}
}

type Option func(opt *SessionOption)

func IDBuilder(id ID) Option {
	return func(opt *SessionOption) {
		opt.ID = id
	}
}

func In(in int) Option {
	return func(opt *SessionOption) {
		opt.in = in
	}
}

func Out(out int) Option {
	return func(opt *SessionOption) {
		opt.out = out
	}
}

func HBInterval(hbInterval time.Duration) Option {
	return func(opt *SessionOption) {
		opt.hbInterval = hbInterval
	}
}

func ReadBuffer(rb int) Option {
	return func(opt *SessionOption) {
		opt.rBuffer = rb
	}
}

func WriteBuffer(wb int) Option {
	return func(opt *SessionOption) {
		opt.wBuffer = wb
	}
}

func WriteTime(wt time.Duration) Option {
	return func(opt *SessionOption) {
		opt.wTime = wt
	}
}

func HandShakeTime(hst time.Duration) Option {
	return func(opt *SessionOption) {
		opt.hsTime = hst
	}
}

func ReadLimit(rLimit int64) Option {
	return func(opt *SessionOption) {
		opt.rLimit = rLimit
	}
}
