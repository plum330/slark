package sf

import "github.com/zeromicro/go-zero/core/syncx"

// golang sf lib  exists panic & exit -> customized

type SingleFlight struct {
	sf syncx.SingleFlight
}

func NewSingFlight() *SingleFlight {
	return &SingleFlight{sf: syncx.NewSingleFlight()}
}

func (s *SingleFlight) Do(key string, fn func() (any, error)) (any, error) {
	return s.sf.Do(key, fn)
}
