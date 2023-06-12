package etcd

import (
	"context"
	"time"
)

type option struct {
	ctx   context.Context
	ns    string
	ttl   time.Duration
	retry int
}

type Option func(*option)
