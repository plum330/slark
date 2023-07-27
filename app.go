package slark

import (
	"context"
	"github.com/go-slark/slark/registry"
	"github.com/go-slark/slark/transport"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	servers  []transport.Server
	signals  []os.Signal
	registry registry.Registry
	name     string
	version  string
}

type Option func(*App)

func Server(srv ...transport.Server) Option {
	return func(app *App) {
		app.servers = srv
	}
}

func Signal(signals ...os.Signal) Option {
	return func(app *App) {
		app.signals = signals
	}
}

func Registry(r registry.Registry) Option {
	return func(app *App) {
		app.registry = r
	}
}

func Name(name string) Option {
	return func(app *App) {
		app.name = name
	}
}

func Version(ver string) Option {
	return func(app *App) {
		app.version = ver
	}
}

func NewApp(opts ...Option) *App {
	app := &App{
		signals: []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV},
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func (a *App) Run() error {
	c := make(chan os.Signal, 1)
	eg, ctx := errgroup.WithContext(context.TODO())
	wg := sync.WaitGroup{}
	for _, server := range a.servers {
		wg.Add(1)
		s := server
		eg.Go(func() error {
			wg.Done()
			return s.Start()
		})

		eg.Go(func() error {
			<-c
			cx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return s.Stop(cx)
		})
	}
	wg.Wait()
	svc, err := a.service()
	if err != nil {
		return err
	}
	if a.registry != nil {
		cx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		err = a.registry.Register(cx, svc)
		if err != nil {
			return err
		}
	}
	signal.Notify(c, a.signals...)
	<-c
	close(c)

	if a.registry != nil {
		cx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()
		return a.registry.Unregister(cx, svc)
	}
	return eg.Wait()
}

func (a *App) service() (*registry.Service, error) {
	var err error
	u := &url.URL{}
	for _, srv := range a.servers {
		ep, ok := srv.(transport.Endpoint)
		if !ok {
			continue
		}
		u, err = ep.Endpoint()
		if err != nil {
			return nil, err
		}
	}
	svc := &registry.Service{
		ID:       uuid.New().String(),
		Name:     a.name,
		Version:  a.version,
		Endpoint: u.String(),
	}
	return svc, nil
}
