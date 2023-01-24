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
	Server []transport.Server
}

func NewApp(srv ...transport.Server) *App {
	return &App{
		Server: srv,
	}
}

func (a *App) Run() error {
	c := make(chan os.Signal, 1)
	eg, ctx := errgroup.WithContext(context.TODO())
	for _, server := range a.Server {
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

	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	<-c
	close(c)
	return eg.Wait()
}
