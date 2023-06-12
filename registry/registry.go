package registry

import "context"

type Service struct {
	ID       string
	Name     string
	Endpoint []string
}

type Registry interface {
	Register(ctx context.Context, svc *Service) error
	Deregister(ctx context.Context, svc *Service) error
}

type Discovery interface {
	List(ctx context.Context, name string) ([]*Service, error)
	//Stop() error
}

type Watcher interface {
	//List() ([]*Service, error)
	//Stop() error
}
