package grpc

import (
	"context"
	"github.com/go-slark/slark/errors"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"net/url"
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

// TODO
func (p *parser) update(svc []*registry.Service) {
	addresses := make([]resolver.Address, 0, len(svc))
	for _, s := range svc {
		u, err := url.Parse(s.Endpoint)
		if err != nil {
			continue
		}
		var address string
		if u.Scheme == utils.Discovery {
			address = u.Host
		}
		addr := resolver.Address{
			ServerName: s.Name,
			//BalancerAttributes 字段可以用来保存负载均衡策略所使用的信息，比如权重信息
			BalancerAttributes: attributes.New("attributes", s),
			Addr:               address,
		}
		addresses = append(addresses, addr)
	}
	_ = p.cc.UpdateState(resolver.State{Addresses: addresses})
}
