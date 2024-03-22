package breaker

import "sync"

// breaker: 服务过载保护 & 服务弹性 & 防止雪崩

type Breaker interface {
	Allow() error
	Fail(reason string)
	Succeed()
}

type Breakers struct {
	bre map[string]Breaker
	l   sync.RWMutex
	f   func() Breaker
}

type Option func(breakers *Breakers)

func WithBreaker(f func() Breaker) Option {
	return func(b *Breakers) {
		b.f = f
	}
}

func NewBreaker(opts ...Option) *Breakers {
	bre := &Breakers{
		bre: make(map[string]Breaker),
		l:   sync.RWMutex{},
		f: func() Breaker {
			return NewGoogleBreaker()
		},
	}
	for _, opt := range opts {
		opt(bre)
	}
	return bre
}

func (b *Breakers) Fetch(name string) Breaker {
	b.l.RLock()
	bre, ok := b.bre[name]
	if ok {
		b.l.RUnlock()
		return bre
	}
	b.l.RUnlock()
	b.l.Lock()
	defer b.l.Unlock()
	b.bre[name] = b.f()
	return b.bre[name]
}
