package algo

import (
	"context"
	"errors"
	"github.com/go-slark/slark/transport/grpc/balancer/node"
	"math/rand"
	"time"
)

type wrr struct {
	r *rand.Rand
}

func NewWRRBuilder() node.Builder {
	return &node.BalancerBuilder{Picker: &wrr{r: rand.New(rand.NewSource(time.Now().UnixNano()))}}
}

func (w *wrr) Pick(_ context.Context, nodes []*node.Node) (*node.Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available node")
	}

	newNodes := make([]*node.Node, 0)
	for _, n := range nodes {
		for i := 0; i < n.Weight; i++ {
			newNodes = append(newNodes, n)
		}
	}
	return newNodes[w.r.Intn(len(newNodes))], nil
}
