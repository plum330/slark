package env

import (
	"context"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/encoding/json"
	"os"
	"strings"
)

type Env struct {
	prefix []string
	ctx    context.Context
	cancel context.CancelFunc
}

type Option func(*Env)

func Prefix(prefix ...string) Option {
	return func(e *Env) {
		e.prefix = prefix
	}
}

func New(opts ...Option) *Env {
	ctx, cancel := context.WithCancel(context.Background())
	e := &Env{
		prefix: []string{"slark_"},
		ctx:    ctx,
		cancel: cancel,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Env) Load() ([]byte, error) {
	mp := make(map[string]any)
	envs := os.Environ()
	for _, env := range envs {
		var key, value string
		str := strings.SplitN(env, "=", 2)
		key = str[0]
		if len(str) > 1 {
			value = str[1]
		}
		prefix, match := e.match(env)
		if match && len(prefix) != len(key) {
			key = strings.TrimPrefix(strings.TrimPrefix(key, prefix), "_")
			if len(key) > 0 {
				mp[key] = value
			}
		}
	}
	return encoding.GetCodec(json.Name).Marshal(mp)
}

func (e *Env) match(str string) (string, bool) {
	for _, prefix := range e.prefix {
		if strings.HasPrefix(str, prefix) {
			return prefix, true
		}
	}
	return "", false
}

func (e *Env) Watch() <-chan struct{} {
	return e.ctx.Done()
}

func (e *Env) Close() error {
	e.cancel()
	return nil
}

func (e *Env) Format() string {
	return json.Name
}
