package metadata

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/metadata"
	"github.com/go-slark/slark/transport"
)

func Metadata(pt middleware.PeerType) middleware.Middleware {
	w := metadata.New()
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				ok    bool
				trans transport.Transporter
				md    metadata.Metadata
			)
			carrier := trans.ReqCarrier()
			if pt == middleware.Server {
				trans, ok = transport.FromServerContext(ctx)
				if !ok {
					return handler(ctx, req)
				}
				md = metadata.Metadata{}
				for _, key := range carrier.Keys() {
					if !w.HasPrefix(key) {
						continue
					}
					for _, value := range carrier.Values(key) {
						md.Add(key, value)
					}
				}
				ctx = metadata.NewMetadataContext(ctx, md)
			} else if pt == middleware.Client {
				trans, ok = transport.FromClientContext(ctx)
				if !ok {
					return handler(ctx, req)
				}
				md, ok = metadata.FromMetadataContext(ctx)
				if !ok {
					return handler(ctx, req)
				}
				for _, key := range carrier.Keys() {
					if !w.HasPrefix(key) {
						continue
					}
					for _, value := range carrier.Values(key) {
						carrier.Add(key, value)
					}
				}
			}
			return handler(ctx, req)
		}
	}
}
