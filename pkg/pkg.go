package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stat"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	LogName         = "log-dumper"
	TraceID         = "x-trace-id"
	SpanID          = "x-span-id"
	Authorization   = "x-authorization"
	Header          = "x-header"
	Token           = "x-token"
	Claims          = "x-jwt"
	UserAgent       = "User-Agent"
	ContentEncoding = "Content-Encoding"
	Method          = "x-method"
	Path            = "x-path"
	Code            = "x-code"
	RequestVars     = "x-request-vars"
	Extension       = "x-extension"

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
	return "", errors.New("scheme not found")
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

func ParseScheme(endpoints []string) (map[string]string, error) {
	mp := make(map[string]string)
	for _, endpoint := range endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		mp[u.Port()] = u.Scheme
	}
	return mp, nil
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

func bitOne(n int64) []int {
	bits := make([]int, 64)
	for i := 0; i < 64; i++ {
		if (n>>i)&1 == 1 {
			bits[i] = 1
		}
	}
	return bits
}

func Filter[T any](ss []T, n int64) []T {
	v := make([]T, 0, len(ss))
	bits := bitOne(n)
	for index, bit := range bits {
		if bit == 1 {
			v = append(v, ss[index])
		}
	}
	return v
}

func Read(fn string) (string, error) {
	text, err := os.ReadFile(fn)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(text)), nil
}

type LineReadOption struct {
	space  bool
	blank  bool
	prefix string
}

type LineReadOpt func(option *LineReadOption)

func WithSpace(space bool) LineReadOpt {
	return func(o *LineReadOption) {
		o.space = space
	}
}

func WithBlank(blank bool) LineReadOpt {
	return func(o *LineReadOption) {
		o.blank = blank
	}
}

func WithPrefix(prefix string) LineReadOpt {
	return func(o *LineReadOption) {
		o.prefix = prefix
	}
}

func ReadLines(fn string, opts ...LineReadOpt) ([]string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	o := &LineReadOption{}
	for _, opt := range opts {
		opt(o)
	}
	lines := make([]string, 0)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if !o.space {
			line = strings.TrimSpace(line)
		}
		if o.blank && len(line) == 0 {
			continue
		}
		if len(o.prefix) > 0 && strings.HasPrefix(line, o.prefix) {
			continue
		}
		lines = append(lines, line)
	}
	return lines, s.Err()
}

func init() {
	// shielding stat log
	stat.DisableLog()
}
