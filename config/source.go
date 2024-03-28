package config

import "io"

type Source interface {
	Load() ([]byte, error)
	Watch() <-chan struct{}
	io.Closer
}
