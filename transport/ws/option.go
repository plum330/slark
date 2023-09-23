package ws

import (
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"time"
)

type ServerOption func(*Server)

func Logger(l logger.Logger) ServerOption {
	return func(s *Server) {
		s.logger = l
	}
}

func WithNetwork(network string) ServerOption {
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

func ConnOpt(opts ...Option) ServerOption {
	return func(server *Server) {
		for _, opt := range opts {
			opt(server.opt)
		}
	}
}

func Handle(handlers ...middleware.HTTPMiddleware) ServerOption {
	return func(server *Server) {
		server.handlers = handlers
	}
}

type Option func(opt *ConnOption)

func WithIn(in int) Option {
	return func(opt *ConnOption) {
		opt.in = in
	}
}

func WithOut(out int) Option {
	return func(opt *ConnOption) {
		opt.out = out
	}
}

func WithHBInterval(hbInterval time.Duration) Option {
	return func(opt *ConnOption) {
		opt.hbInterval = hbInterval
	}
}

func WithReadBuffer(rb int) Option {
	return func(opt *ConnOption) {
		opt.rBuffer = rb
	}
}

func WithWriteBuffer(wb int) Option {
	return func(opt *ConnOption) {
		opt.wBuffer = wb
	}
}

func WithWriteTime(wt time.Duration) Option {
	return func(opt *ConnOption) {
		opt.wTime = wt
	}
}

func WithHandShakeTime(hst time.Duration) Option {
	return func(opt *ConnOption) {
		opt.hsTime = hst
	}
}

func WithReadLimit(rLimit int64) Option {
	return func(opt *ConnOption) {
		opt.rLimit = rLimit
	}
}

func WithCloseTime(tm time.Duration) Option {
	return func(opt *ConnOption) {
		if tm != 0 {
			opt.closeTime = tm
		}
	}
}
