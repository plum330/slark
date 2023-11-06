package ws

import (
	"fmt"
	"github.com/go-slark/slark/middleware/cors"
	"net/http"
	"testing"
)

func TestWebsocket(t *testing.T) {
	srv := NewServer(
		Address("0.0.0.0:9090"),
		Path("/ws"),
		Handlers(cors.CORS()),
		Before(func(writer http.ResponseWriter, request *http.Request) (interface{}, error) {
			return nil, nil
			//return errors.BadRequest("请求错误", "REQUEST_ERROR")
		}),
		After(func(s *Session) error {
			return nil
		}),
	)
	srv.Handler(func(s *Session) {
		for {
			msg, e := s.Receive()
			if e != nil {
				fmt.Printf("receive msg fail, id:%s, err:%+v\n", s.ID(), e)
				return
			}

			fmt.Printf("receive msg, id:%s, result:%s\n", s.ID(), msg.Payload)
			msg.Payload = []byte(s.ID())
			e = s.Send(msg)
			if e != nil {
				fmt.Printf("send msg fail, id:%s, err:%+v\n", s.ID(), e)
				return
			}
			fmt.Printf("send msg succ, id:%s, msg: %s\n", s.ID(), msg.Payload)
		}
	})

	err := srv.Start()
	if err != nil {
		fmt.Println("websocket start fail, err:", err)
	}
}
