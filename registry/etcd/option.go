package etcd

import (
	"context"
)

type option struct {
	ctx   context.Context
	ns    string
	ttl   int64
	retry int
}

type Option func(*option)

func Context(ctx context.Context) Option {
	return func(o *option) {
		o.ctx = ctx
	}
}

func Namespace(ns string) Option {
	return func(o *option) {
		o.ns = ns
	}
}

func TTL(ttl int64) Option {
	return func(o *option) {
		o.ttl = ttl
	}
}

func Retry(r int) Option {
	return func(o *option) {
		o.retry = r
	}
}
