package algo

import (
	"context"
	"errors"
	"github.com/go-slark/slark/transport/grpc/balancer/node"
	"sync"
)

type wrr struct {
	l      sync.Mutex
	weight map[string]int64
}

func NewWRRBuilder() node.Builder {
	return &node.BalancerBuilder{
		Picker:          &wrr{weight: map[string]int64{}},
		WeightedBuilder: &node.Plain{},
	}
}

func (w *wrr) Pick(_ context.Context, nodes []node.WeightedNode) (node.WeightedNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available node")
	}
	var (
		ew, tw, cw, hw int64
		addr           string
		hn             node.WeightedNode
	)
	w.l.Lock()
	for _, n := range nodes {
		ew = n.Weight() // effective weight
		addr = n.Address()
		cw = w.weight[addr] + ew
		w.weight[addr] = cw // current weight
		tw += ew            // total weight
		if hn == nil || hw < cw {
			hw = cw // hit weight
			hn = n  // hit node
		}
	}
	w.weight[hn.Address()] = hw - tw
	w.l.Unlock()
	return hn, nil
}
