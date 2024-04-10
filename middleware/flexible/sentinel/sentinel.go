package sentinel

import (
	"context"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg/flexible/sentinel"
	"github.com/go-slark/slark/transport"
)

func Sentinel(pt middleware.PeerType, opts ...sentinel.Option) middleware.Middleware {
	s, _ := sentinel.New("", opts...)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				trans       transport.Transporter
				ok          bool
				trafficType base.TrafficType
				resType     base.ResourceType
			)
			if pt == middleware.Server {
				trafficType = base.Inbound
				trans, ok = transport.FromServerContext(ctx)
			} else if pt == middleware.Client {
				trafficType = base.Outbound
				trans, ok = transport.FromClientContext(ctx)
			}
			if !ok || !s.Exists(trans.Operate()) {
				return handler(ctx, req)
			}
			kind := trans.Kind()
			if kind == transport.GRPC {
				resType = base.ResTypeRPC
			} else if kind == transport.HTTP {
				resType = base.ResTypeWeb
			}
			entry, e := api.Entry(
				trans.Operate(),
				api.WithResourceType(resType),
				api.WithTrafficType(trafficType),
			)
			if e != nil {
				// TODO fallback
				return nil, e
			}
			defer entry.Exit()
			rsp, err := handler(ctx, req)
			if err != nil {
				api.TraceError(entry, err)
			}
			return rsp, nil
		}
	}
}
