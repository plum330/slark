package grpc

import (
	"context"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// server

func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return middleware.HandleMiddleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}, s.mw...)(ctx, req)
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
						ch <- p
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
				l.Lock()
				defer l.Unlock()
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

func UnaryServerTracID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.Pairs()
			}

			requestID := md[pkg.TraceID]
			if len(requestID) > 0 {
				ctx = context.WithValue(ctx, pkg.TraceID, requestID[0])
				return handler(ctx, req)
			}

			ctx = context.WithValue(ctx, pkg.TraceID, pkg.BuildRequestID())
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

func UnaryClientTraceID() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		value := ctx.Value(pkg.TraceID)
		requestID, ok := value.(string)
		if !ok || len(requestID) == 0 {
			requestID = pkg.BuildRequestID()
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.Pairs()
		}
		md[pkg.TraceID] = []string{requestID}
		return invoker(metadata.NewOutgoingContext(ctx, md), method, req, resp, cc, opts...)
	}
}

func DialOpts() []grpc.DialOption {
	retryOps := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithPerRetryTimeout(time.Second * 2),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(time.Second/2, 0.2)),
	}
	retry := grpc_retry.UnaryClientInterceptor(retryOps...)
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(retry, UnaryClientTimeout(3*time.Second)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
	}
	return opts
}
