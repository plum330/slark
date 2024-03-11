package http

import (
	"context"
	"github.com/go-slark/slark/errors"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

/*
go install github.com/rakyll/hey
设置max_conn = 100
hey -z 1s -c 90 -q 1 'http://localhost:8080/ping' (压测90个并发,执行1s)
hey -z 1s -c 110 -q 1 'http://localhost:8080/ping' (压测110个并发,执行1s)
*/

func TestServer(t *testing.T) {
	srv := NewServer(Builtin(0x63))
	r := NewRouter(srv)
	r.Handle(http.MethodGet, "/ping", func(ctx *Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	srv.Start()
}

func TestBreaker(t *testing.T) {
	srv := NewServer(Builtin(0x63))
	r := NewRouter(srv)
	rr := rand.NewSource(time.Now().UnixMilli())
	r.Handle(http.MethodGet, "/ping", func(ctx *Context) error {
		x, err := ctx.Handle(func(ctx context.Context, req interface{}) (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			i := rr.Int63()
			if i&0x1 == 1 {
				//return nil
			}
			return nil, errors.ServerUnavailable("xx-err-msg", "xx-err-reason")
		})(ctx.Context(), nil)
		if err != nil {
			return err
		}
		return ctx.Result(x)
	})
	srv.Start()
}
