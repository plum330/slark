package http

import (
	"github.com/gin-gonic/gin"
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
	srv := NewServer(Builtin(0x14b), MaxConn(100))
	srv.engine.Handle(http.MethodGet, "/ping", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusOK, "success")
	})
	srv.Start()
}
