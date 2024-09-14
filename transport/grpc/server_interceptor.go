package grpc

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/opentelemetry/trace"
	"github.com/go-slark/slark/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		trans := &Transport{
			operation: info.FullMethod,
			req:       Carrier(md),
			rsp:       Carrier{},
		}
		ctx = transport.NewServerContext(ctx, trans)
		var cancel context.CancelFunc
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}
		return middleware.ComposeMiddleware(s.mws...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		})(ctx, req)
	}
}

func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		trans := &Transport{
			req: Carrier(md),
			rsp: Carrier{},
		}
		ctx = transport.NewServerContext(ctx, trans)
		_, err := middleware.ComposeMiddleware(s.mws...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, handler(srv, &ssWrapper{ctx: ctx, ServerStream: ss})
		})(ctx, nil)
		return err
	}
}

func ServerTimeout(timeout time.Duration) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			var (
				resp interface{}
				err  error
				l    sync.Mutex
			)
			done := make(chan struct{})
			ch := make(chan interface{}, 1)
			go func() {
				defer func() {
					if p := recover(); p != nil {
						ch <- fmt.Sprintf("%+v\n\n%s", p, strings.TrimSpace(string(debug.Stack())))
					}
				}()

				l.Lock()
				defer l.Unlock()
				resp, err = handler(ctx, req)
				close(done)
			}()

			select {
			case p := <-ch:
				panic(p)
			case <-done:
				l.Lock()
				defer l.Unlock()
				return resp, err
			case <-ctx.Done():
				err = ctx.Err()
				if err == context.Canceled {
					err = status.Error(codes.Canceled, err.Error())
				} else if err == context.DeadlineExceeded {
					err = status.Error(codes.DeadlineExceeded, err.Error())
				}
				return nil, err
			}
		}
	}
}

func ServerTraceID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			requestID := md[utils.TraceID]
			if len(requestID) > 0 {
				ctx = context.WithValue(ctx, utils.TraceID, requestID[0])
			}
			return handler(ctx, req)
		}
	}
}

func ServerAuthZ() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			token := md[utils.Token]
			if len(token) > 0 {
				str, err := strconv.Unquote(token[0])
				if err != nil {
					return nil, err
				}
				ctx = context.WithValue(ctx, utils.Token, str)
			}
			return handler(ctx, req)
		}
	}
}

type ssWrapper struct {
	grpc.ServerStream
	ctx    context.Context
	rMsgID int // received msg id
	sMsgID int // send msg id
}

func (w *ssWrapper) SendMsg(m interface{}) error {
	w.sMsgID++
	trace.MessageSent.Event(w.ctx, w.sMsgID, m)
	return w.ServerStream.SendMsg(m)
}

func (w *ssWrapper) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}
	w.rMsgID++
	trace.MessageReceived.Event(w.ctx, w.rMsgID, m)
	return nil
}

func (w *ssWrapper) Context() context.Context {
	return w.ctx
}
