package utils

import (
	"context"
	"encoding/json"
	"github.com/go-slark/slark/errors"
	"github.com/google/uuid"
	"net"
)

const (
	LogName       = "log-dumper"
	RayID         = "x-request-id"
	SpanID        = "x-span-id"
	Authorization = "x-authorization"
	Header        = "x-header"
	Token         = "x-token"
	Claims        = "x-jwt"
	UserAgent     = "User-Agent"
	Target        = "x-target"
	Method        = "x-method"
	Path          = "x-path"
	Code          = "x-code"
	RequestVars   = "x-request-vars"
	Extension     = "x-extension"
	Filter        = "x-filter"

	XForwardedMethod = "X-Forwarded-Method"
	XForwardedURI    = "X-Forwarded-Uri"
	XForwardedIP     = "X-Forwarded-For"

	ContentType = "Content-Type"
	Accept      = "Accept"
	Application = "application"

	Discovery = "discovery"
	Weight    = "weight"
)

func BuildRequestID() string {
	return uuid.New().String()
}

func ParseToken(ctx context.Context, v interface{}) error {
	token, ok := ctx.Value(Token).(string)
	if !ok {
		return errors.BadRequest(errors.TokenError, errors.TokenError)
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
func FilterValidIP() ([]net.IP, error) {
	is, err := net.Interfaces()
	if err != nil {
		return nil, err
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
	return ips, nil
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
