package tracing

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/peer"
	"net"
	"strconv"
	"strings"
)

func attributes(ctx context.Context, fullMethod string) []attribute.KeyValue {
	_, methodAttrs := parseFullMethod(fullMethod)
	peerAttrs := peerAttr(peerFromCtx(ctx))
	attrs := make([]attribute.KeyValue, 0, len(methodAttrs)+len(peerAttrs))
	attrs = append(attrs, methodAttrs...)
	attrs = append(attrs, peerAttrs...)
	return attrs
}

func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
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

	var attrs []attribute.KeyValue
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

func peerAttr(addr string) []attribute.KeyValue {
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

	var attrs []attribute.KeyValue
	if ip := net.ParseIP(host); ip != nil {
		attrs = []attribute.KeyValue{
			semconv.NetSockPeerAddr(host),
			semconv.NetSockPeerPort(port),
		}
	} else {
		attrs = []attribute.KeyValue{
			semconv.NetPeerName(host),
			semconv.NetPeerPort(port),
		}
	}
	return attrs
}

func httpAttributes(operation string) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	ss := strings.Split(operation, " ")
	if len(ss) != 2 {
		return attrs
	}
	attrs = append(attrs, semconv.HTTPMethod(ss[0]))
	attrs = append(attrs, semconv.HTTPTarget(ss[1]))
	return attrs
}
