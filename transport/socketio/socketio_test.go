package socketio

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"testing"
)

func TestSocketIO(t *testing.T) {
	srv := NewServer(Addr("localhost:8089"))
	srv.OnConnect("/", func(conn socketio.Conn) error {
		fmt.Println("connected:", conn.ID())
		conn.Emit("/msg", "test msg")
		return nil
	})
	_ = srv.Start()
}
