package config

type Source interface {
	Load() ([]byte, error)
	Watch() <-chan struct{}
	Close() error
	Format() string
}
