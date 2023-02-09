package trace

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
)

func SpanInfo(fullMethod, addr string) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{semconv.RPCSystemKey.String("grpc")}
	name, attr := ParseFullMethod(fullMethod)
	attrs = append(attrs, attr...)
	attrs = append(attrs, PeerAttr(addr)...)
	return name, attrs
}

func ParseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		// Invalid format, does not follow `/package.service/method`.
		return name, []attribute.KeyValue(nil)
	}

	var attrs []attribute.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}
	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}
	return name, attrs
}

func PeerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}

	if len(host) == 0 {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetSockPeerAddrKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

func parseAddr(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil {
		return ""
	}
	return p.Addr.String()
}
