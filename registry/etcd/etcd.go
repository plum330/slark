package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/pkg/retry"
	"github.com/go-slark/slark/registry"
	"go.etcd.io/etcd/client/v3"
	"time"
)

type Registry struct {
	client *clientv3.Client
	lease  clientv3.Lease
	opt    *option
}

func New(cfg clientv3.Config, opts ...Option) *Registry {
	opt := &option{
		ctx:   context.Background(),
		ns:    "/default",
		ttl:   10,
		retry: 5,
	}
	for _, o := range opts {
		o(opt)
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		panic(fmt.Sprintf("create etcd err:%+v", err))
	}
	return &Registry{
		client: client,
		opt:    opt,
	}
}

func (r *Registry) Register(ctx context.Context, svc *registry.Service) error {
	key := fmt.Sprintf("/%s/%s/%s", r.opt.ns, svc.Name, svc.ID)
	value, _ := json.Marshal(svc)
	if r.lease != nil {
		// release lease resource
		_ = r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	leaseID, err := r.put(ctx, key, string(value))
	if err != nil {
		return err
	}
	go r.keepAlive(r.opt.ctx, leaseID, key, string(value))
	return nil
}

func (r *Registry) put(ctx context.Context, key, value string) (clientv3.LeaseID, error) {
	grant, err := r.lease.Grant(ctx, int64(time.Duration(r.opt.ttl)*time.Second))
	if err != nil {
		return 0, err
	}
	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *Registry) keepAlive(ctx context.Context, leaseID clientv3.LeaseID, key, value string) {
	ch, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		leaseID = 0
	}

	for {
		if leaseID == 0 {
			err = retry.NewOption(
				retry.Retry(r.opt.retry), retry.Delay(500*time.Millisecond), retry.MaxJitter(500*time.Millisecond)).Retry(func() error {
				e := ctx.Err()
				if e != nil {
					return e
				}

				// non-blocking
				ic := make(chan clientv3.LeaseID, 1)
				ec := make(chan error, 1)
				cx, cancel := context.WithTimeout(ctx, 3*time.Second)
				go func() {
					defer cancel()
					id, pe := r.put(cx, key, value)
					if pe != nil {
						ec <- pe
					} else {
						ic <- id
					}
				}()

				select {
				case <-cx.Done():
					return errors.InternalServer("time out", "TIME_OUT")
				case e = <-ec:
					return e
				case leaseID = <-ic:
				}

				ch, err = r.client.KeepAlive(ctx, leaseID)
				return err
			})
			if err != nil {
				return
			}

			if _, ok := <-ch; !ok {
				return
			}
		}

		select {
		case _, ok := <-ch:
			if !ok {
				if ctx.Err() != nil {
					return
				}
				// retry
				leaseID = 0
				continue
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Registry) Unregister(ctx context.Context, svc *registry.Service) error {
	key := fmt.Sprintf("/%s/%s/%s", r.opt.ns, svc.Name, svc.ID)
	_, err := r.client.Delete(ctx, key)
	if r.lease != nil {
		_ = r.lease.Close()
	}
	return err
}

//func (r *Registry) Discover(ctx context.Context, name string) ([]*registry.Service, error) {
//	key := fmt.Sprintf("/%s/%s", r.opt.ns, name)
//	cx, _ := context.WithCancel(ctx) // TODO
//	watcher := clientv3.NewWatcher(r.client)
//	watcher.Watch(cx, key, clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
//	err := watcher.RequestProgress(cx)
//	if err != nil {
//		return nil, err
//	}
//	rsp, err := clientv3.NewKV(r.client).Get(cx, key, clientv3.WithPrefix())
//	if err != nil {
//		return nil, err
//	}
//
//	svc := make([]*registry.Service, 0, len(rsp.Kvs))
//	for _, kv := range rsp.Kvs {
//		s := &registry.Service{}
//		err = json.Unmarshal(kv.Value, s)
//		if err != nil {
//			return nil, err
//		}
//		if s.Name != name {
//			continue
//		}
//		svc = append(svc, s)
//	}
//	return svc, nil
//}

//type watcher struct {
//	key         string
//	ctx         context.Context
//	cancel      context.CancelFunc
//	client      *clientv3.Client
//	watchChan   clientv3.WatchChan
//	watcher     clientv3.Watcher
//	kv          clientv3.KV
//	first       bool
//	serviceName string
//}

//func (w *watcher) List() ([]*registry.Service, error) {
//	// 是否需要区分first
//	if w.first {
//		item, err := w.getService()
//		w.first = false
//		return item, err
//	}
//
//	select {
//	case <-w.ctx.Done():
//		return nil, w.ctx.Err()
//	case watchResp, ok := <-w.watchChan:
//		if !ok || watchResp.Err() != nil {
//			time.Sleep(time.Second)
//			err := w.reWatch()
//			if err != nil {
//				return nil, err
//			}
//		}
//		return w.getService()
//	}
//}

//func (w *watcher) getService() ([]*registry.Service, error) {
//	rsp, err := w.kv.Get(w.ctx, w.key, clientv3.WithPrefix())
//	if err != nil {
//		return nil, err
//	}
//
//	svc := make([]*registry.Service, 0, len(rsp.Kvs))
//	for _, kv := range rsp.Kvs {
//		s := &registry.Service{}
//		err = json.Unmarshal(kv.Value, s)
//		if err != nil {
//			return nil, err
//		}
//		if s.Name != w.serviceName {
//			continue
//		}
//		svc = append(svc, s)
//	}
//	return svc, nil
//}

//func (w *watcher) reWatch() error {
//	w.watcher.Close()
//	w.watcher = clientv3.NewWatcher(w.client)
//	w.watchChan = w.watcher.Watch(w.ctx, w.key, clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
//	return w.watcher.RequestProgress(w.ctx)
//}

//func (w *watcher) Stop() error {
//	w.cancel()
//	return w.watcher.Close()
//}
