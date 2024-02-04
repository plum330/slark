package grpc

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
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

func (s *Server) interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = context.WithValue(ctx, utils.Method, info.FullMethod)
		return middleware.ComposeMiddleware(s.mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		})(ctx, req)
	}
}

func UnaryServerTimeout(timeout time.Duration) middleware.Middleware {
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

func UnaryServerTrace() middleware.Middleware {
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

func UnaryServerAuthorize() middleware.Middleware {
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

//func ServerTrace(opts ...tracing.Option) middleware.Middleware {
//	tracer := tracing.NewTracing(trace.SpanKindServer, opts...)
//	return func(handler middleware.Handler) middleware.Handler {
//		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
//			md, ok := metadata.FromIncomingContext(ctx)
//			if !ok {
//				md = metadata.MD{}
//			}
//			name, _ := ctx.Value(utils.Method).(string)
//			//name, attr := SpanInfo(name, parseAddr(ctx))
//			ctx, span := tracer.Start(ctx, name, &tracing.Carrier{MD: md}, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attr...))
//			//defer tracer.Stop(ctx, span, rsp, err)
//			defer span.End()
//			return handler(ctx, req)
//		}
//	}
//}
