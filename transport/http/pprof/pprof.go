package pprof

import (
	"context"
	"github.com/go-slark/slark/errors"
	"net/http"
	_ "net/http/pprof"
)

type Server struct {
	*http.Server
}

func NewServer(addr string) *Server {
	return &Server{
		Server: &http.Server{
			Addr: addr,
		},
	}
}

func (s *Server) Start() error {
	if len(s.Addr) == 0 {
		return nil
	}
	return s.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if len(s.Addr) == 0 {
		return nil
	}
	err := s.Shutdown(ctx)
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
