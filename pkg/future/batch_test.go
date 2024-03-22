package future

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	b := NewBatch(WithExec(func(ctx context.Context, m map[string][]any) {
		for _, vv := range m {
			for _, v := range vv {
				fmt.Println(v)
			}
		}
	}), WithSharding(func(key string) int {
		return len(key)
	}))
	b.Add("0000", "9999")
	b.Run()
	time.Sleep(5 * time.Second)
	b.Stop()
}
