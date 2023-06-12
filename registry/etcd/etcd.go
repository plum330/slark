package etcd

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/registry"
	"go.etcd.io/etcd/client/v3"
	"time"
)

type Registry struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
	opt    *option
}

func New(cfg clientv3.Config, opts ...Option) *Registry {
	opt := &option{
		ctx:   context.Background(),
		ns:    "/default",
		ttl:   10 * time.Second,
		retry: 10,
	}
	for _, o := range opts {
		o(opt)
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		panic(fmt.Sprintf("new etcd client fail, err:%+v", err))
	}
	return &Registry{
		client: client,
		kv:     clientv3.NewKV(client),
		opt:    opt,
	}
}

func (r *Registry) Register(ctx context.Context, svc *registry.Service) error {
	return nil
}

func (r *Registry) Deregister(ctx context.Context, svc *registry.Service) error {
	return nil
}

func (r *Registry) List(ctx context.Context, name string) ([]*registry.Service, error) {
	return nil, nil
}
