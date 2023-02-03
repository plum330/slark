package bind

import (
	"fmt"
	"net/http"
	"testing"
)

type Test struct {
	Name string
	Age  int
}

func TestBind(t *testing.T) {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		test := &Test{}
		err := Bind(r, test)
		if err != nil {
			fmt.Println("bind:", err)
			return
		}

		fmt.Printf("test:%+v\n", test)
	})
	err := http.ListenAndServe(":7070", nil)
	if err != nil {
		fmt.Println("listen and server:", err)
	}
}
