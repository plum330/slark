package ws

import (
	"time"
)

type ServerOption func(*Server)

func WithNetwork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func WithAddress(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithPath(path string) ServerOption {
	return func(s *Server) {
		s.path = path
	}
}

func WithConnOption(opts ...Option) ServerOption {
	return func(server *Server) {
		for _, opt := range opts {
			opt(server.ConnOption)
		}
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
