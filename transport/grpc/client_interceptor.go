package grpc

import (
	"context"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
)

func unaryClientInterceptor(opt *option) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		trans := &Transport{
			operation: method,
			req:       Carrier(md),
			rsp:       Carrier{},
			filters:   opt.filters,
		}
		ctx = transport.NewClientContext(ctx, trans)
		if opt.tm > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opt.tm)
			defer cancel()
		}
		_, err := middleware.ComposeMiddleware(opt.mws...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = metadata.NewOutgoingContext(ctx, md)
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		})(ctx, req)
		return err
	}
}

func streamClientInterceptor(opt *option) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		trans := &Transport{
			operation: method,
			req:       Carrier(md),
			rsp:       Carrier{},
			filters:   opt.filters,
		}
		ctx = transport.NewClientContext(ctx, trans)
		rsp, err := middleware.ComposeMiddleware(opt.mws...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = metadata.NewOutgoingContext(ctx, md)
			return streamer(ctx, desc, cc, method, opts...)
		})(ctx, nil)
		return rsp.(grpc.ClientStream), err
	}
}

func ClientTraceID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			value := ctx.Value(utils.TraceID)
			requestID, ok := value.(string)
			if !ok || len(requestID) == 0 {
				requestID = utils.BuildRequestID()
			}
			ctx = metadata.AppendToOutgoingContext(ctx, utils.TraceID, requestID)
			return handler(ctx, req)
		}
	}
}

func ClientAuthZ() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			token, ok := ctx.Value(utils.Token).(string)
			if ok {
				ctx = metadata.AppendToOutgoingContext(ctx, utils.Token, strconv.QuoteToASCII(token))
			}
			return handler(ctx, req)
		}
	}
}

type clientStreamWrapper struct {
	grpc.ClientStream
	rMsgID int
	sMsgID int
}

func (w *clientStreamWrapper) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err != nil {
		return err
	}
	w.rMsgID++
	return nil
}

func (w *clientStreamWrapper) SendMsg(m interface{}) error {
	w.sMsgID++
	return w.ClientStream.SendMsg(m)
}
