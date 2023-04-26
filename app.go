package slark

import (
	"context"
	"github.com/go-slark/slark/transport"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	servers []transport.Server
	signals []os.Signal
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
	for _, server := range a.servers {
		s := server
		eg.Go(func() error {
			return s.Start()
		})

		eg.Go(func() error {
			<-c
			cx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			return s.Stop(cx)
		})
	}

	signal.Notify(c, a.signals...)
	<-c
	close(c)
	return eg.Wait()
}
