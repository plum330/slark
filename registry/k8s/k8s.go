package k8s

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/registry"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type Registry struct {
	name     string
	ns       string
	interval time.Duration
}

type Option func(*Registry)

func Interval(interval time.Duration) Option {
	return func(r *Registry) {
		r.interval = interval
	}
}

func Namespace(ns string) Option {
	return func(r *Registry) {
		r.ns = ns
	}
}

func Name(name string) Option {
	return func(r *Registry) {
		r.name = name
	}
}

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{
		ns:       "/default",
		name:     "svc",
		interval: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Registry) Register(ctx context.Context, svc *registry.Service) error {
	return nil
}

func (r *Registry) Unregister(ctx context.Context, svc *registry.Service) error {
	return nil
}

func (r *Registry) Discover(ctx context.Context, name string) (registry.Watcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(config)
	inf := informers.NewSharedInformerFactoryWithOptions(cs, r.interval,
		informers.WithNamespace(r.ns),
		informers.WithTweakListOptions(func(options *metaV1.ListOptions) {
			options.FieldSelector = "name=" + r.name
		}))
	in := inf.Core().V1().Endpoints()
	notify := make(chan struct{}, 1)
	_, _ = in.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("uuuuuuuuuuuuuuu")
			_, ok := obj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			notify <- struct{}{}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("wwwwwwwwwwwwwwww")
			oEndpoints, ok := oldObj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			nEndpoints, ok := newObj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			if oEndpoints.ResourceVersion == nEndpoints.ResourceVersion {
				return
			}
			notify <- struct{}{}
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("vvvvvvvvvvvv")
			_, ok := obj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			notify <- struct{}{}
		},
	})
	go inf.Start(context.Background().Done())
	w := &watcher{
		clientSet: cs,
		notify:    notify,
		ns:        r.ns,
		name:      r.name,
	}
	return w, nil
}

type watcher struct {
	clientSet *kubernetes.Clientset
	ns        string
	name      string
	notify    chan struct{}
}

func (w *watcher) List() ([]*registry.Service, error) {
	endpoints, err := w.clientSet.CoreV1().Endpoints(w.ns).Get(context.Background(), w.name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}
	svc := make([]*registry.Service, 0, len(endpoints.Subsets))
	for index, set := range endpoints.Subsets {
		for _, addr := range set.Addresses {
			s := &registry.Service{
				ID:       "",
				Name:     addr.Hostname,
				Version:  "",
				Endpoint: addr.IP,
			}
			svc = append(svc, s)
			fmt.Printf("k8s list svc, no:%d, svc:%+v\n", index, s)
		}
	}
	return svc, nil
}

func (w *watcher) Stop() error {
	return nil
}
