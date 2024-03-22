package breaker

import (
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/stat"
	"sync"
)

type GoogleBreaker struct {
	once sync.Once
	breaker.Breaker
	breaker.Promise
}

func NewGoogleBreaker() *GoogleBreaker {
	stat.DisableLog()
	return &GoogleBreaker{Breaker: breaker.NewBreaker()}
}

func (g *GoogleBreaker) Allow() error {
	promise, err := g.Breaker.Allow()
	g.once.Do(func() {
		g.Promise = promise
	})
	return err
}

func (g *GoogleBreaker) Fail(reason string) {
	g.Promise.Reject(reason)
}

func (g *GoogleBreaker) Succeed() {
	g.Promise.Accept()
}
