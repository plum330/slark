package balancer

import (
	"github.com/go-slark/slark/registry"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"sync"
	"time"
)

type wrr struct {
	c []balancer.SubConn
	r *rand.Rand
	l sync.Mutex
}

func NewWRRBuilder() Builder {
	return &wrr{}
}

func (w *wrr) Build() Picker {
	return &wrr{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (w *wrr) Update(scs map[balancer.SubConn]base.SubConnInfo) {
	w.c = make([]balancer.SubConn, 0, len(scs))
	for k, v := range scs {
		svc, _ := v.Address.BalancerAttributes.Value("attributes").(*registry.Service)
		// TODO default mix/max weight
		weight, _ := svc.Metadata["weight"].(int)
		for i := 0; i < weight; i++ {
			w.c = append(w.c, k)
		}
	}
}

func (w *wrr) Pick() balancer.SubConn {
	w.l.Lock()
	sc := w.c[w.r.Intn(len(w.c))]
	w.l.Unlock()
	return sc
}
