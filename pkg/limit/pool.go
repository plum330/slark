package limit

import "errors"

type Pool struct {
	worker chan struct{}
}

func NewPool(size int) *Pool {
	return &Pool{worker: make(chan struct{}, size)}
}

func (p *Pool) Use() bool {
	select {
	case p.worker <- struct{}{}:
		return true
	default:
		return false
	}
}

func (p *Pool) Back() error {
	select {
	case <-p.worker:
		return nil
	default:
		return errors.New("discard worker")
	}
}
