package routine

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type Service struct{}

func (s *Service) Do() {
	fmt.Println("+++++++")
}

type Consumer struct{}

func (c *Consumer) Do() {
	fmt.Println("=======")
}

func TestGroup(t *testing.T) {
	g := NewGroup()
	g.Append(&Service{}, &Consumer{})
	g.Do()
}

func TestGo(t *testing.T) {
	for i := 0; i < 5; i++ {
		j := i
		GoSafe(context.TODO(), func() {
			fmt.Println(j)
		})
	}
	time.Sleep(10 * time.Second)
}
