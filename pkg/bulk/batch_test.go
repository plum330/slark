package bulk

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBulk(t *testing.T) {
	v := make([]int, 0)
	var l sync.Mutex
	bt := NewBatch(func(vs []any) {
		l.Lock()
		defer l.Unlock()
		v = append(v, len(vs))
	}, BatchInterval(3*time.Second))
	for i := 0; i < 10; i++ {
		bt.Submit(i)
	}
	time.Sleep(5 * time.Second)
	for index, value := range v {
		fmt.Printf("index:%d, value:%v\n", index, value)
	}
}
