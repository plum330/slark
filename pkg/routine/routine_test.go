package routine

import (
	"fmt"
	"testing"
)

type Service struct{}

func (s *Service) Start() {
	fmt.Println("+++++++")
}

func TestGroup(t *testing.T) {
	g := NewGroup()
	g.Append(&Service{})
	g.Start()
}
