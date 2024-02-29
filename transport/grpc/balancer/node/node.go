package node

import (
	"context"
	"errors"
	"google.golang.org/grpc/balancer"
	"sync"
)

type WrappedNode struct {
	Addr    string
	Weight  *int64
	SubConn balancer.SubConn
}

type Node interface {
	InitialWeight() *int64
	Address() string
}

func (w *WrappedNode) Address() string {
	return w.Addr
}

func (w *WrappedNode) InitialWeight() *int64 {
	return w.Weight
}

type WeightedNode interface {
	Node
	Weight() int64
	//Unwrap() Node
}

type Set struct {
	nodes   []WeightedNode
	l       sync.RWMutex
	picker  Picker
	builder WeightedBuilder
}

type Picker interface {
	Pick(ctx context.Context, nodes []WeightedNode) (WeightedNode, error)
}

type Builder interface {
	Build() Balancer
}

type Filter func(ctx context.Context, nodes []Node) []Node

type Balancer interface {
	Save(nodes []Node)
	Pick(ctx context.Context, fs ...Filter) (Node, error)
}

func (s *Set) Save(nodes []Node) {
	wn := make([]WeightedNode, 0, len(nodes))
	for _, node := range nodes {
		wn = append(wn, s.builder.Build(node))
	}
	s.l.Lock()
	s.nodes = wn
	s.l.Unlock()
}

func (s *Set) Pick(ctx context.Context, filters ...Filter) (Node, error) {
	var nodes []WeightedNode
	s.l.RLock()
	nodes = s.nodes
	s.l.RUnlock()
	if len(nodes) == 0 {
		return nil, errors.New("no available node")
	}
	ns := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		ns = append(ns, node)
	}
	for _, filter := range filters {
		ns = filter(ctx, ns)
	}
	cn := make([]WeightedNode, len(ns))
	for idx, n := range ns {
		cn[idx] = n.(WeightedNode)
	}
	if len(filters) == 0 {
		cn = nodes
	}
	if len(cn) == 0 {
		return nil, errors.New("no available node")
	}
	// balance algo
	return s.picker.Pick(ctx, cn)
}

type BalancerBuilder struct {
	Picker
	WeightedBuilder
}

func (b *BalancerBuilder) Build() Balancer {
	return &Set{
		picker:  b.Picker,
		builder: b.WeightedBuilder,
	}
}
