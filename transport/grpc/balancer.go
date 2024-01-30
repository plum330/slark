package grpc

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"sync"
	"time"
)

const Balance = "random"

type rPickerBuilder struct{}

func init() {
	balancer.Register(base.NewBalancerBuilder(
		Balance,
		&rPickerBuilder{},
		base.Config{HealthCheck: true},
	))
}

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
