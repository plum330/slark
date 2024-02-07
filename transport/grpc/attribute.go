package grpc

import (
	"context"
	attr "go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/peer"
	"net"
	"strconv"
	"strings"
)

func attribute(ctx context.Context, fullMethod string) (string, []attr.KeyValue) {
	name, methodAttrs := parseFullMethod(fullMethod)
	peerAttrs := peerAttr(peerFromCtx(ctx))
	attrs := make([]attr.KeyValue, 0, len(methodAttrs)+len(peerAttrs))
	attrs = append(attrs, methodAttrs...)
	attrs = append(attrs, peerAttrs...)
	return name, attrs
}

func parseFullMethod(fullMethod string) (string, []attr.KeyValue) {
	if !strings.HasPrefix(fullMethod, "/") {
		// Invalid format, does not follow `/package.service/method`.
		return fullMethod, nil
	}
	name := fullMethod[1:]
	pos := strings.LastIndex(name, "/")
	if pos < 0 {
		// Invalid format, does not follow `/package.service/method`.
		return name, nil
	}
	service, method := name[:pos], name[pos+1:]

	var attrs []attr.KeyValue
	if service != "" {
		attrs = append(attrs, semconv.RPCService(service))
	}
	if method != "" {
		attrs = append(attrs, semconv.RPCMethod(method))
	}
	return name, attrs
}

func peerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

func peerAttr(addr string) []attr.KeyValue {
	host, p, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}

	if host == "" {
		host = "127.0.0.1"
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return nil
	}

	var attrs []attr.KeyValue
	if ip := net.ParseIP(host); ip != nil {
		attrs = []attr.KeyValue{
			semconv.NetSockPeerAddr(host),
			semconv.NetSockPeerPort(port),
		}
	} else {
		attrs = []attr.KeyValue{
			semconv.NetPeerName(host),
			semconv.NetPeerPort(port),
		}
	}
	return attrs
}
