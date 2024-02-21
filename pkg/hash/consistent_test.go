package hash

import (
	"testing"
)

func TestConsistent(t *testing.T) {
	h := New()
	nodes := []string{"127.0.0.1", "126.0.0.1", "125.0.0.1"}
	for _, node := range nodes {
		h.Add(node)
	}
	node := h.Fetch("125.0.0.1" + "0")
	t.Logf("node:%s", node)
}
