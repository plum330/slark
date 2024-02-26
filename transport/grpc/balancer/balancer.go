package balancer

import (
	utils "github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/registry"
	"github.com/go-slark/slark/transport/grpc/balancer/algo"
	"github.com/go-slark/slark/transport/grpc/balancer/node"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"strconv"
)

const LoadBalancer = "load_balancer"

var pickerBuilder = &builder{Builder: algo.NewRandomBuilder()}

func SetBuilder(builder node.Builder) {
	pickerBuilder.Builder = builder
}

func init() {
	balancer.Register(base.NewBalancerBuilder(
		LoadBalancer,
		pickerBuilder,
		base.Config{HealthCheck: true},
	))
}

type builder struct {
	node.Builder
}

func (b *builder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	nodes := make([]*node.Node, 0, len(info.ReadySCs))
	for sc, si := range info.ReadySCs {
		svc, ok := si.Address.Attributes.Value("attributes").(*registry.Service)
		n := &node.Node{
			Addr:    si.Address.Addr,
			SubConn: sc,
		}
		if ok {
			n.Service = svc
			weight, _ := svc.Metadata[utils.Weight]
			n.Weight, _ = strconv.Atoi(weight)
		}
		nodes = append(nodes, n)
	}
	p := &picker{
		Balancer: b.Builder.Build(),
	}
	p.Save(nodes)
	return p
}

type picker struct {
	node.Balancer
}

func (p *picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	filters, _ := info.Ctx.Value(utils.Filter).([]node.Filter)
	n, err := p.Balancer.Pick(info.Ctx, filters...)
	if err != nil {
		return balancer.PickResult{}, err
	}
	result := balancer.PickResult{
		SubConn: n.SubConn,
		Done: func(info balancer.DoneInfo) {
			// TODO
		},
	}
	return result, nil
}
