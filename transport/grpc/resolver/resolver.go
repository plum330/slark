package resolver

import (
	"context"
	"github.com/go-slark/slark/errors"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/endpoint"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type parser struct {
	watcher  registry.Watcher
	ctx      context.Context
	cancel   context.CancelFunc
	cc       resolver.ClientConn
	ss       Subset
	size     int
	insecure bool
}

func (p *parser) ResolveNow(opts resolver.ResolveNowOptions) {}

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
}

func (p *parser) update(svc []*registry.Service) {
	mp := map[string]struct{}{}
	set := make([]*registry.Service, 0, len(svc))
	var ok bool
	// filter
	for _, s := range svc {
		addr, err := endpoint.ParseValidAddr(s.Endpoint, endpoint.Scheme("grpc", p.insecure))
		if err != nil {
			continue
		}
		_, ok = mp[addr]
		if ok {
			continue
		}
		mp[addr] = struct{}{}
		set = append(set, s)
	}
	if p.ss != nil && p.size > 0 {
		set = p.ss.Subset(set, p.size)
	}
	addresses := make([]resolver.Address, 0, len(svc))
	for _, s := range set {
		addr, _ := endpoint.ParseValidAddr(s.Endpoint, endpoint.Scheme("grpc", p.insecure))
		address := resolver.Address{
			ServerName: s.Name,
			Attributes: attributes.New(utils.ServiceRegistry, s),
			Addr:       addr,
		}
		addresses = append(addresses, address)
	}
	if len(addresses) == 0 {
		return
	}
	_ = p.cc.UpdateState(resolver.State{Addresses: addresses})
}
