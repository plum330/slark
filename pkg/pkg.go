package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net"
	"net/url"
	"strconv"
)

const (
	LogName       = "log-dumper"
	TraceID       = "x-trace-id"
	SpanID        = "x-span-id"
	Authorization = "x-authorization"
	Header        = "x-header"
	Token         = "x-token"
	Claims        = "x-jwt"
	UserAgent     = "User-Agent"
	Method        = "x-method"
	Path          = "x-path"
	Code          = "x-code"
	RequestVars   = "x-request-vars"
	Extension     = "x-extension"

	XForwardedMethod = "X-Forwarded-Method"
	XForwardedURI    = "X-Forwarded-Uri"
	XForwardedIP     = "X-Forwarded-For"

	ContentType = "Content-Type"
	Accept      = "Accept"
	Application = "application"

	Discovery       = "discovery"
	Weight          = "weight"
	ServiceRegistry = "service-registry"
)

func BuildRequestID() string {
	return uuid.New().String()
}

func ParseToken(ctx context.Context, v interface{}) error {
	token, ok := ctx.Value(Token).(string)
	if !ok {
		return errors.New("invalid token")
	}
	return json.Unmarshal([]byte(token), v)
}

func SnakeCase(s string) string {
	l := len(s)
	b := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			b = append(b, '_')
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return string(b)
}

func Scheme(scheme string, insecure bool) string {
	if !insecure {
		return scheme
	}
	return scheme + "s"
}

func ParseValidAddr(addr []string, scheme string) (string, error) {
	for _, v := range addr {
		u, err := url.Parse(v)
		if err != nil {
			return "", err
		}
		if u.Scheme == scheme {
			return u.Host, nil
		}
	}
	return "", nil
}

func ParseAddr(ln net.Listener, address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil && ln == nil {
		return "", err
	}
	if ln != nil {
		tcpAddr, ok := ln.Addr().(*net.TCPAddr)
		if !ok {
			return "", errors.New("parse addr error")
		}
		port = strconv.Itoa(tcpAddr.Port)
	}

	is, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	index := int(^uint(0) >> 1)
	ips := make([]net.IP, 0)
	for _, i := range is {
		if (i.Flags & net.FlagUp) == 0 {
			continue
		}
		if i.Index >= index && len(ips) != 0 {
			continue
		}

		addr, e := i.Addrs()
		if e != nil {
			continue
		}
		for _, a := range addr {
			var ip net.IP
			switch at := a.(type) {
			case *net.IPAddr:
				ip = at.IP
			case *net.IPNet:
				ip = at.IP
			default:
				continue
			}

			ipBytes := net.ParseIP(ip.String())
			if !ipBytes.IsGlobalUnicast() || ipBytes.IsInterfaceLocalMulticast() {
				continue
			}
			index = i.Index
			ips = append(ips, ip)
			if ip.To4() != nil {
				break
			}
		}
	}
	var host string
	if len(ips) != 0 {
		host = net.JoinHostPort(ips[len(ips)-1].String(), port)
	}
	return host, nil
}

func Delete[T comparable](ss []T, elem T) []T {
	var index int
	for _, s := range ss {
		if s != elem {
			ss[index] = s
			index++
		}
	}
	return ss[:index]
}

func BitOne(n int64) []int {
	bits := make([]int, 64)
	for i := 0; i < 64; i++ {
		if (n>>i)&1 == 1 {
			bits[i] = 1
		}
	}
	return bits
}
