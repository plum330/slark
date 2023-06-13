package grpc

import (
	"context"
	"errors"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

type parser struct {
	watcher registry.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
	cc      resolver.ClientConn
}

func (p *parser) ResolveNow(opts resolver.ResolveNowOptions) {
	// TODO
}

func (p *parser) Close() {
	p.cancel()
	_ = p.watcher.Stop()
}

func (p *parser) watch() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		svc, err := p.watcher.List()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			time.Sleep(time.Second)
			continue
		}
		p.update(svc)
	}
}

func (p *parser) update(svc []*registry.Service) {
	addrs := make([]resolver.Address, 0, len(svc))
	for _, s := range svc {
		addr := resolver.Address{
			ServerName: s.Name,
			//Attributes 字段可以用来保存负载均衡策略所使用的信息，比如权重信息
			Addr: "",
		}
		addrs = append(addrs, addr)
	}
	err := p.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
	}
}
