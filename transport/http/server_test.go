package http

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/transport/http/handler"
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
	srv := NewServer(Enable(0x63))
	r := NewRouter(srv)
	r.Handle(http.MethodGet, "/ping", func(ctx *Context) error {
		time.Sleep(1 * time.Millisecond)
		fmt.Println("++++++++++++")
		return nil
	})
	srv.Start()
}

func TestMetric(t *testing.T) {
	srv := NewServer(Enable(0x67))
	r := NewRouter(srv)
	r.Handle(http.MethodGet, "/ping", func(ctx *Context) error {
		x, err := ctx.Handle(func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})(ctx.Context(), nil)
		if err != nil {
			return err
		}
		return ctx.Result(x)
	})
	srv.Start()
}

func TestBreaker(t *testing.T) {
	srv := NewServer(Enable(0x63))
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

func TestRedirect(t *testing.T) {
	srv := NewServer(Enable(0x63), Handlers(handler.CORS(), handler.Redirect(&handler.Redirecting{
		URL:         "/redirect",
		RedirectURL: "https://www.baidu.com",
		Code:        301,
	})))
	r := NewRouter(srv)
	r.Handle(http.MethodGet, "/redirect", func(ctx *Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	srv.Start()
}

func TestURI(t *testing.T) {
	srv := NewServer(Enable(0x63))
	r := NewRouter(srv)
	r.Handle(http.MethodGet, "/info/:name", func(ctx *Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	r.Handle(http.MethodGet, "/qname", func(ctx *Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	srv.Start()
}
