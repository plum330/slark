package balancer

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
	"sync"
	"time"
)

type random struct {
	c []balancer.SubConn
	r *rand.Rand
	l sync.Mutex
}

func NewRandomBuilder() Builder {
	return &random{}
}

func (r *random) Build() Picker {
	return &random{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (r *random) Update(scs map[balancer.SubConn]base.SubConnInfo) {
	r.c = make([]balancer.SubConn, 0, len(scs))
	for sc := range scs {
		r.c = append(r.c, sc)
	}
}

func (r *random) Pick() balancer.SubConn {
	r.l.Lock()
	sc := r.c[r.r.Intn(len(r.c))]
	r.l.Unlock()
	return sc
}
