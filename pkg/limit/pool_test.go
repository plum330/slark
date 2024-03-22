package limit

import (
	"fmt"
	"testing"
)

func TestPool(t *testing.T) {
	pool := NewPool(2)
	result := pool.Use()
	if result {
		fmt.Println("--------")
	}
	result = pool.Use()
	if result {
		fmt.Println("--------")
	}
	result = pool.Use()
	if !result {
		fmt.Println("++++")
	}
}
