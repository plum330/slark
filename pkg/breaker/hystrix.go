package breaker

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/rs/xid"
	"time"
)

type Hystrix struct {
	*hystrix.CircuitBreaker
}

func NewHystrix() *Hystrix {
	bre, _, _ := hystrix.GetCircuit(xid.New().String())
	return &Hystrix{CircuitBreaker: bre}
}

func (h *Hystrix) Allow() error {
	allow := h.CircuitBreaker.AllowRequest()
	if allow {
		return nil
	}
	return errors.New("breaker open request forbidden")
}

func (h *Hystrix) Fail(reason string) {
	_ = h.CircuitBreaker.ReportEvent([]string{"failure"}, time.Now(), time.Second)
}

func (h *Hystrix) Succeed() {
	_ = h.CircuitBreaker.ReportEvent([]string{"success"}, time.Now(), time.Second)
}
