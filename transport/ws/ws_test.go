package ws

import (
	"fmt"
	"net/http"
	"testing"
)

func TestWebsocket(t *testing.T) {
	srv := NewServer(WithAddress("0.0.0.0:9090"), WithPath("/ws"))
	srv.Handler(func(w http.ResponseWriter, r *http.Request) {
		session, err := srv.NewSession(w, r)
		if err != nil {
			fmt.Println("new session fail, err:", err)
			return
		}

		go func() {
			for {
				msg, e := session.Receive()
				if e != nil {
					fmt.Printf("receive msg fail, id:%s, err:%+v\n", session.ID(), e)
					return
				}

				fmt.Printf("receive msg, id:%s, result:%s\n", session.ID(), msg.Payload)
				msg.Payload = []byte(session.ID())
				e = session.Send(msg)
				if e != nil {
					fmt.Printf("send msg fail, id:%s, err:%+v\n", session.ID(), e)
					return
				}
				fmt.Printf("send msg succ, id:%s, msg: %s\n", session.ID(), msg.Payload)
			}
		}()
	})

	err := srv.Start()
	if err != nil {
		fmt.Println("websocket start fail, err:", err)
	}
}
