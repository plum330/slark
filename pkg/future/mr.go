package future

import (
	"context"
	"github.com/zeromicro/go-zero/core/mr"
)

type Parallel[T, U, V any] struct {
	ctx      context.Context
	cancel   func(error)
	producer mr.GenerateFunc[T]
	splitter mr.MapperFunc[T, U]
	merger   mr.ReducerFunc[U, V]
	worker   int
}

type Option[T, U, V any] func(*Parallel[T, U, V])

func Workers[T, U, V any](worker int) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.worker = worker
	}
}

func Context[T, U, V any](ctx context.Context) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.ctx = ctx
	}
}

func Splitter[T, U, V any](splitter mr.MapperFunc[T, U]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.splitter = splitter
	}
}

func Merger[T, U, V any](merge mr.ReducerFunc[U, V]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.merger = merge
	}
}

func Producer[T, U, V any](producer mr.GenerateFunc[T]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.producer = producer
	}
}

func NewParallel[T, U, V any](opts ...Option[T, U, V]) *Parallel[T, U, V] {
	parallel := &Parallel[T, U, V]{
		worker: 16,
		ctx:    context.TODO(),
	}
	for _, opt := range opts {
		opt(parallel)
	}
	return parallel
}

func (p *Parallel[T, U, V]) Do() (V, error) {
	return mr.MapReduce(p.producer, p.splitter, p.merger, mr.WithContext(p.ctx), mr.WithWorkers(p.worker))
}

func Exec(fs ...func() error) error {
	return mr.Finish(fs...)
}

func VoidExec(fs ...func()) {
	mr.FinishVoid(fs...)
}
