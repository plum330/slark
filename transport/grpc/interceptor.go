package grpc

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

// server

func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
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

func ServerOpts() []grpc.ServerOption {
	option := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	return option
}

// client

func UnaryClientTimeout(defaultTime time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			defaultTimeout := defaultTime
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		}
		if cancel != nil {
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientTrace() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		value := ctx.Value(utils.RayID)
		requestID, ok := value.(string)
		if !ok || len(requestID) == 0 {
			requestID = utils.BuildRequestID()
		}
		ctx = metadata.AppendToOutgoingContext(ctx, utils.RayID, requestID)
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}

func UnaryClientAuthorize() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token, ok := ctx.Value(utils.Token).(string)
		if ok {
			ctx = metadata.AppendToOutgoingContext(ctx, utils.Token, strconv.QuoteToASCII(token))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Client) unaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = context.WithValue(ctx, utils.Target, map[string]string{method: cc.Target()})
		_, err := middleware.ComposeMiddleware(c.mw...)(func(ctx context.Context, req interface{}) (interface{}, error) {
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		})(ctx, req)
		return err
	}
}
