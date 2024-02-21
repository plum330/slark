package sf

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestSF(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	g := NewGroup()
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			data, _, _ := g.Do("test", func() (interface{}, error) {
				time.Sleep(2 * time.Second)
				return rand.Int63(), nil
			})
			fmt.Println("data:", data)
		}()
	}
	wg.Wait()
}
