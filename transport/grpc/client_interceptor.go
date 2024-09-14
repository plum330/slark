package grpc

import (
	"context"
	"errors"
	"github.com/go-slark/slark/middleware"
	utils "github.com/go-slark/slark/pkg"
	tracing "github.com/go-slark/slark/pkg/opentelemetry/trace"
	"github.com/go-slark/slark/transport"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
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
			s, err := streamer(ctx, desc, cc, method, opts...)
			if err != nil {
				return nil, err
			}
			cs := wrapClientStream(ctx, s, desc)

			// trace
			go func() {
				span := trace.SpanFromContext(ctx)
				err = <-cs.finished
				if err != nil {
					st, o := status.FromError(err)
					if o {
						span.SetStatus(codes.Error, st.Message())
						span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(st.Code())))
					} else {
						span.SetStatus(codes.Error, err.Error())
					}
				} else {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(gcodes.OK)))
				}
			}()
			return cs, nil
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

type streamEventType int

type streamEvent struct {
	Type streamEventType
	Err  error
}

const (
	receiveEndEvent streamEventType = iota
	errorEvent
)

type clientStreamWrapper struct {
	grpc.ClientStream
	rMsgID   int
	sMsgID   int
	finished chan error
	desc     *grpc.StreamDesc
	done     chan struct{}
	events   chan streamEvent
}

func wrapClientStream(ctx context.Context, s grpc.ClientStream, desc *grpc.StreamDesc) *clientStreamWrapper {
	events := make(chan streamEvent)
	done := make(chan struct{})
	finished := make(chan error)

	go func() {
		defer close(done)
		for {
			select {
			case event := <-events:
				switch event.Type {
				case receiveEndEvent:
					finished <- nil
					return
				case errorEvent:
					finished <- event.Err
					return
				}
			case <-ctx.Done():
				finished <- ctx.Err()
				return
			}
		}
	}()

	return &clientStreamWrapper{
		ClientStream: s,
		desc:         desc,
		events:       events,
		done:         done,
		finished:     finished,
	}
}

func (w *clientStreamWrapper) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err != nil {
		if errors.Is(err, io.EOF) {
			w.sendStreamEvent(receiveEndEvent, nil)
		} else {
			w.sendStreamEvent(errorEvent, err)
		}
	} else {
		if !w.desc.ServerStreams {
			w.sendStreamEvent(receiveEndEvent, nil)
		} else {
			w.rMsgID++
			tracing.MessageReceived.Event(w.Context(), w.rMsgID, m)
		}
	}
	return err
}

func (w *clientStreamWrapper) SendMsg(m interface{}) error {
	w.sMsgID++
	tracing.MessageSent.Event(w.Context(), w.sMsgID, m)
	err := w.ClientStream.SendMsg(m)
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}
	return err
}

func (w *clientStreamWrapper) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}
	return md, err
}

func (w *clientStreamWrapper) CloseSend() error {
	err := w.ClientStream.CloseSend()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}
	return err
}

func (w *clientStreamWrapper) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-w.done:
	case w.events <- streamEvent{Type: eventType, Err: err}:
	}
}
