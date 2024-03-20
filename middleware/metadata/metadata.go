package metadata

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/transport"
)

type metadata map[string][]string

func (m metadata) add(key, value string) {
	m[key] = append(m[key], value)
}

type metadataContext struct{}

func NewMetadataContext(ctx context.Context, md metadata) context.Context {
	return context.WithValue(ctx, metadataContext{}, md)
}

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			trans, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			md := metadata{}
			carrier := trans.ReqCarrier()
			for _, key := range carrier.Keys() {
				for _, value := range carrier.Values(key) {
					md.add(key, value)
				}
			}
			ctx = NewMetadataContext(ctx, md)
			return handler(ctx, req)
		}
	}
}
