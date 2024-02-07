package grpc

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	tracing "github.com/go-slark/slark/pkg/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
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
		if strings.HasPrefix(info.FullMethod, "/") {
			pos := strings.LastIndex(info.FullMethod[1:], "/")
			if pos >= 0 {
				ctx = context.WithValue(ctx, utils.Method, info.FullMethod[1:][pos+1:])
				ctx = context.WithValue(ctx, utils.Path, info.FullMethod[1:][:pos])
			}
		}
		return middleware.ComposeMiddleware(s.mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		})(ctx, req)
	}
}

func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		cx := ss.Context()
		metadata.FromIncomingContext(cx)
		_, err := middleware.ComposeMiddleware(s.mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, handler(srv, &ssWrapper{ctx: cx, ServerStream: ss})
		})(cx, nil)
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

func ServerRayID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			requestID := md[utils.RayID]
			if len(requestID) > 0 {
				ctx = context.WithValue(ctx, utils.RayID, requestID[0])
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

func UnaryServerTrace(opts ...tracing.Option) grpc.UnaryServerInterceptor {
	tracer := tracing.NewTracer(trace.SpanKindClient, opts...)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		name, attrs := attribute(ctx, info.FullMethod)
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(tracer.Kind()),
			trace.WithAttributes(attrs...),
		}
		ctx, span := tracer.Start(ctx, name, &tracing.Carrier{MD: md}, opt...)
		defer span.End()
		resp, err := handler(ctx, req)
		s, _ := status.FromError(err)
		if err != nil {
			span.SetStatus(statusCode(s))
		}
		span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(s.Code().String()))
		return resp, err
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
	// TODO trace event
	return w.ServerStream.SendMsg(m)
}

func (w *ssWrapper) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}
	w.rMsgID++
	// TODO trace event
	return nil
}

func (w *ssWrapper) Context() context.Context {
	return w.ctx
}

func StreamServerTrace(opts ...tracing.Option) grpc.StreamServerInterceptor {
	tracer := tracing.NewTracer(trace.SpanKindClient, opts...)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		c := ss.Context()
		md, ok := metadata.FromIncomingContext(c)
		if !ok {
			md = metadata.MD{}
		}
		name, attrs := attribute(c, info.FullMethod)
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(tracer.Kind()),
			trace.WithAttributes(attrs...),
		}
		cx, span := tracer.Start(c, name, &tracing.Carrier{MD: md}, opt...)
		defer span.End()
		err := handler(srv, &ssWrapper{ServerStream: ss, ctx: cx})
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(statusCode(s))
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(s.Code().String()))
		} else {
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(codes.OK.String()))
		}
		return err
	}
}
