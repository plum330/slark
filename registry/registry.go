package registry

import "context"

type Service struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Endpoint []string `json:"endpoint"`
}

type Registry interface {
	Register(ctx context.Context, svc *Service) error
	Unregister(ctx context.Context, svc *Service) error
}

type Discovery interface {
	Discover(ctx context.Context, name string) ([]*Service, error)
}

//type Watcher interface {
//	List() ([]*Service, error)
//	Stop() error
//}
