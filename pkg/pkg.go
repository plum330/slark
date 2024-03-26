package utils

import (
	"bufio"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stat"
	"os"
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

func ParseToken(token string, v any) error {
	return json.Unmarshal([]byte(token), v)
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
