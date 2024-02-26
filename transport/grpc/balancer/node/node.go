package node

import (
	"context"
	"errors"
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/balancer"
	"sync"
)

type Node struct {
	Addr    string
	Weight  int
	Service *registry.Service
	SubConn balancer.SubConn
}

// Set default
type Set struct {
	nodes  []*Node
	l      sync.RWMutex
	picker Picker
}

type Picker interface {
	Pick(ctx context.Context, nodes []*Node) (*Node, error)
}

type Builder interface {
	Build() Balancer
}

type Filter func(ctx context.Context, nodes []*Node) []*Node

type Balancer interface {
	Save(nodes []*Node)
	Pick(ctx context.Context, fs ...Filter) (*Node, error)
}

func (s *Set) Save(nodes []*Node) {
	s.l.Lock()
	defer s.l.Unlock()
	s.nodes = nodes
}

func (s *Set) Pick(ctx context.Context, filters ...Filter) (*Node, error) {
	nodes := make([]*Node, 0)
	s.l.RLock()
	nodes = s.nodes
	s.l.RUnlock()
	if len(nodes) == 0 {
		return nil, errors.New("no available node")
	}
	cNodes := make([]*Node, 0, len(nodes))
	for _, filter := range filters {
		cNodes = filter(ctx, nodes)
	}
	if len(filters) == 0 {
		cNodes = nodes
	}
	if len(cNodes) == 0 {
		return nil, errors.New("no available node")
	}
	// balance algo
	return s.picker.Pick(ctx, cNodes)
}

type BalancerBuilder struct {
	Picker
}

func (b *BalancerBuilder) Build() Balancer {
	return &Set{
		picker: b.Picker,
	}
}
