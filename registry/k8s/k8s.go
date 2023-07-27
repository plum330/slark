package k8s

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/registry"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"strings"
	"time"
)

type Registry struct {
	interval time.Duration
	token    string
}

type Option func(*Registry)

func Interval(interval time.Duration) Option {
	return func(r *Registry) {
		r.interval = interval
	}
}

func Token(token string) Option {
	return func(r *Registry) {
		r.token = token
	}
}

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{
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

// name(k8s集群中的服务地址) : service-name.namespace.svc.cluster_name:8080

func (r *Registry) Discover(_ context.Context, name string) (registry.Watcher, error) {
	str := strings.FieldsFunc(name, func(r rune) bool {
		return r == ':'
	})
	var (
		port int
		err  error
	)
	if len(str) == 2 {
		port, err = strconv.Atoi(str[1])
		if err != nil {
			return nil, err
		}
	}

	str = strings.FieldsFunc(name, func(r rune) bool {
		return r == '.'
	})
	if len(str) < 2 {
		return nil, errors.InternalServer("k8s target url path invalid", "K8S_TARGET_URL_PATH_INVALID")
	}
	name = str[0]
	ns := str[1]
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config.TLSClientConfig = rest.TLSClientConfig{Insecure: true}
	config.BearerToken = r.token
	config.BearerTokenFile = ""
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	inf := informers.NewSharedInformerFactoryWithOptions(cs, r.interval,
		informers.WithNamespace(ns),
		informers.WithTweakListOptions(func(options *metaV1.ListOptions) {
			options.FieldSelector = "metadata.name=" + name
		}))
	in := inf.Core().V1().Endpoints()
	notify := make(chan struct{}, 1)
	_, err = in.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// TODO -> LOG
			fmt.Println("uuuuuuuuuuuuuuu")
			_, ok := obj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			notify <- struct{}{}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// TODO -> LOG
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
			// TODO -> LOG
			fmt.Println("vvvvvvvvvvvv")
			_, ok := obj.(*coreV1.Endpoints)
			if !ok {
				return
			}
			notify <- struct{}{}
		},
	})
	if err != nil {
		return nil, err
	}
	go inf.Start(context.Background().Done())
	w := &watcher{
		clientSet: cs,
		notify:    notify,
		ns:        ns,
		name:      name,
		port:      port,
	}
	return w, nil
}

type watcher struct {
	clientSet *kubernetes.Clientset
	ns        string
	name      string
	port      int
	notify    chan struct{}
}

func (w *watcher) List() ([]*registry.Service, error) {
	<-w.notify
	endpoints, err := w.clientSet.CoreV1().Endpoints(w.ns).Get(context.Background(), w.name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}
	svc := make([]*registry.Service, 0, len(endpoints.Subsets))
	// meta.name --> len(endpoints.Subsets) == 1
	for _, set := range endpoints.Subsets {
		if len(set.Ports) == 0 {
			break
		}
		if w.port == 0 {
			w.port = int(set.Ports[0].Port)
		}
		for _, addr := range set.Addresses {
			s := &registry.Service{
				Name:     endpoints.Name,
				Version:  endpoints.ResourceVersion,
				Endpoint: fmt.Sprintf("%s:%d", addr.IP, w.port),
			}
			if addr.TargetRef != nil {
				s.ID = string(addr.TargetRef.UID)
			}
			svc = append(svc, s)
			// TODO -> LOG
			fmt.Printf("k8s list svc:%+v\n", s)
		}
	}
	return svc, nil
}

func (w *watcher) Stop() error {
	return nil
}
