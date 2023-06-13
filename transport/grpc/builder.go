package grpc

import (
	"context"
	"github.com/go-slark/slark/errors"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

type builder struct {
	discovery registry.Discovery
	tm        time.Duration
}

func NewBuilder(dis registry.Discovery) resolver.Builder {
	return &builder{
		discovery: dis,
		tm:        10 * time.Second,
	}
}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var (
		err     error
		watcher registry.Watcher
	)
	ch := make(chan struct{}, 1)
	cx, cancel := context.WithCancel(context.Background())
	go func() {
		watcher, err = b.discovery.Discover(cx, strings.TrimPrefix(target.URL.Path, "/"))
		ch <- struct{}{}
	}()

	tm := time.NewTimer(b.tm)
	defer tm.Stop()
	select {
	case <-ch:

	case <-tm.C:
		err = errors.InternalServer("discovery timeout", "DISCOVERY_TIMEOUT")
	}
	if err != nil {
		cancel()
		return nil, err
	}

	p := &parser{
		watcher: watcher,
		cancel:  cancel,
		cc:      cc,
		ctx:     cx,
	}
	go p.start()
	return p, nil
}

func (b *builder) Scheme() string {
	return utils.Discovery
}
