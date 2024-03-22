package metadata

import (
	"context"
	"strings"
)

const (
	MDPrefix       = "x-md-"
	GlobalMDPrefix = "x-md-global-"
	LocalMDPrefix  = "x-md-local-"
)

type Metadata map[string][]string

func (m Metadata) Add(key, value string) {
	if len(key) == 0 {
		return
	}
	key = strings.ToLower(key)
	m[key] = append(m[key], value)
}

type metadataContext struct{}

func NewMetadataContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, metadataContext{}, md)
}

func FromMetadataContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(metadataContext{}).(Metadata)
	return md, ok
}

type Wrapper struct {
	prefix []string
	md     Metadata
}

type Option func(*Wrapper)

func NewWrapper(opts ...Option) *Wrapper {
	w := &Wrapper{
		prefix: []string{MDPrefix},
		md:     nil,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *Wrapper) HasPrefix(key string) bool {
	key = strings.ToLower(key)
	for _, prefix := range w.prefix {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}
