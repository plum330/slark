package ws

import (
	"github.com/go-slark/slark/transport/http/filter"
	"time"
)

type ServerOption func(*Server)

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
			opt(server.ConnOption)
		}
	}
}

func Filter(filters ...filter.Handler) ServerOption {
	return func(server *Server) {
		server.filters = filters
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
