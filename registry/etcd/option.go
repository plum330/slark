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
