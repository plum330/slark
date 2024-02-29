package algo

import (
	"context"
	"errors"
	"github.com/go-slark/slark/transport/grpc/balancer/node"
	"math/rand"
	"time"
)

type random struct {
	r *rand.Rand
}

func NewRandomBuilder() node.Builder {
	return &node.BalancerBuilder{Picker: &random{r: rand.New(rand.NewSource(time.Now().UnixNano()))}}
}

func (r *random) Pick(_ context.Context, nodes []node.WeightedNode) (node.WeightedNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available node")
	}

	return nodes[r.r.Intn(len(nodes))], nil
}
