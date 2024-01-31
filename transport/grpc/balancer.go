package grpc

import (
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"sync"
	"time"
)

const (
	Random = "random"
	WRR    = "weighted_round_robin"
)

func init() {
	balancer.Register(base.NewBalancerBuilder(
		Random,
		&rPickerBuilder{},
		base.Config{HealthCheck: true},
	))
	balancer.Register(base.NewBalancerBuilder(
		WRR,
		&wrrPickerBuilder{},
		base.Config{HealthCheck: true},
	))
}

type rPickerBuilder struct{}

func (r *rPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	rp := &rPicker{
		c: make([]balancer.SubConn, 0, len(info.ReadySCs)),
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for sc := range info.ReadySCs {
		rp.c = append(rp.c, sc)
	}
	return rp
}

type rPicker struct {
	c []balancer.SubConn
	r *rand.Rand
	l sync.Mutex
}

func (r *rPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	r.l.Lock()
	defer r.l.Unlock()
	pr := balancer.PickResult{
		SubConn: r.c[rand.Intn(len(r.c))],
		Done: func(info balancer.DoneInfo) {
			// TODO
		},
	}
	return pr, nil
}

type wrrPickerBuilder struct{}

func (w *wrrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	wp := &wrrPicker{
		c: make([]balancer.SubConn, 0, len(info.ReadySCs)),
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for sc, sci := range info.ReadySCs {
		svc, _ := sci.Address.BalancerAttributes.Value("attributes").(*registry.Service)
		// TODO default mix/max weight
		weight, _ := svc.Metadata["weight"].(int)
		for i := 0; i < weight; i++ {
			wp.c = append(wp.c, sc)
		}
	}
	return wp
}

type wrrPicker struct {
	c []balancer.SubConn
	r *rand.Rand
	l sync.Mutex
}

func (w *wrrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	w.l.Lock()
	defer w.l.Unlock()
	pr := balancer.PickResult{
		SubConn: w.c[rand.Intn(len(w.c))],
		Done: func(info balancer.DoneInfo) {
			// TODO
		},
	}
	return pr, nil
}
