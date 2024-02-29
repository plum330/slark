package resolver

import (
	"github.com/go-slark/slark/registry"
	"math/rand"
)

type Subset interface {
	Subset([]*registry.Service, int) []*registry.Service
}

type Shuffle struct{}

func (s *Shuffle) Subset(set []*registry.Service, size int) []*registry.Service {
	rand.Shuffle(len(set), func(i, j int) {
		set[i], set[j] = set[j], set[i]
	})
	if len(set) <= size {
		return set
	}
	return set[:size]
}
