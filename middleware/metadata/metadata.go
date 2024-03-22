package metadata

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/metadata"
	"github.com/go-slark/slark/transport"
)

func Server() middleware.Middleware {
	w := metadata.NewWrapper()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			trans, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			md := metadata.Metadata{}
			carrier := trans.ReqCarrier()
			for _, key := range carrier.Keys() {
				if !w.HasPrefix(key) {
					continue
				}
				for _, value := range carrier.Values(key) {
					md.Add(key, value)
				}
			}
			ctx = metadata.NewMetadataContext(ctx, md)
			return handler(ctx, req)
		}
	}
}

func Client() middleware.Middleware {
	w := metadata.NewWrapper()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			trans, ok := transport.FromClientContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			reqCarrier := trans.ReqCarrier()
			md, ok := metadata.FromMetadataContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			for key, value := range md {
				if !w.HasPrefix(key) {
					continue
				}
				for _, v := range value {
					reqCarrier.Add(key, v)
				}
			}
			return handler(ctx, req)
		}
	}
}
