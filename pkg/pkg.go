package pkg

import "github.com/google/uuid"

const (
	LogName       = "log-name"
	TraceID       = "x-request-id"
	Authorization = "x-authorization"
	Token         = "x-token"

	Target = "x-target"
	Method = "x-method"
)

func BuildRequestID() string {
	return uuid.New().String()
}

type Config struct {
	Builder   func() string
	RequestId string
}

type Option func(*Config)

func WithBuilder(b func() string) Option {
	return func(cfg *Config) {
		cfg.Builder = b
	}
}

func WithRequestId(requestId string) Option {
	return func(cfg *Config) {
		cfg.RequestId = requestId
	}
}
