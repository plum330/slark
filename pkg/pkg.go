package utils

import (
	"context"
	"encoding/json"
	"github.com/go-slark/slark/errors"
	"github.com/google/uuid"
)

const (
	LogName       = "log-dumper"
	RayID         = "x-request-id"
	Authorization = "x-authorization"
	Token         = "x-token"
	Claims        = "x-jwt"

	Target      = "x-target"
	Method      = "x-method"
	RequestVars = "x-request-vars"

	ContentType = "Content-Type"
	Accept      = "Accept"
	Application = "application"
)

func BuildRequestID() string {
	return uuid.New().String()
}

type Config struct {
	Builder   func() string
	RequestID string
}

type Option func(*Config)

func WithBuilder(b func() string) Option {
	return func(cfg *Config) {
		cfg.Builder = b
	}
}

func WithRequestId(requestID string) Option {
	return func(cfg *Config) {
		cfg.RequestID = requestID
	}
}

func MustParseToken(ctx context.Context, v interface{}) {
	token, ok := ctx.Value(Token).(string)
	if !ok {
		panic(errors.TokenError)
	}
	err := json.Unmarshal([]byte(token), v)
	if err != nil {
		panic(err)
	}
}

func SnakeCase(s string) string {
	l := len(s)
	b := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			b = append(b, '_')
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return string(b)
}
