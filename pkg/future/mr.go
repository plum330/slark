package future

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type (
	Produce[T any]  func(chan<- T)
	Split[T, U any] func(T, Writer[U], func(error))
	Merge[U, V any] func(chan U, Writer[V], func(error))
)

type Writer[T any] interface {
	Write(T)
}

type ec struct {
	once sync.Once
	ch   chan any
}

func (e *ec) write(v any) {
	e.once.Do(func() {
		e.ch <- v
	})
}

type Parallel[T, U, V any] struct {
	ctx      context.Context
	cancel   func(error)
	producer Produce[T]
	splitter Split[T, U]
	merger   Merge[U, V]
	ec       *ec
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

func Splitter[T, U, V any](splitter Split[T, U]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.splitter = splitter
	}
}

func Merger[T, U, V any](merge Merge[U, V]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.merger = merge
	}
}

func Producer[T, U, V any](producer Produce[T]) Option[T, U, V] {
	return func(p *Parallel[T, U, V]) {
		p.producer = producer
	}
}

func NewParallel[T, U, V any](opts ...Option[T, U, V]) *Parallel[T, U, V] {
	parallel := &Parallel[T, U, V]{
		worker: 8,
		ctx:    context.TODO(),
		ec: &ec{
			once: sync.Once{},
			ch:   make(chan any),
		},
	}
	for _, opt := range opts {
		opt(parallel)
	}
	return parallel
}

func (p *Parallel[T, U, V]) produce() chan T {
	input := make(chan T)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				p.ec.write(r)
			}
			close(input)
		}()
		p.producer(input)
	}()
	return input
}

func (p *Parallel[T, U, V]) Do() (V, error) {
	input := p.produce()
	return p.merge(input)
}

func (p *Parallel[T, U, V]) merge(input <-chan T) (V, error) {
	coll := make(chan U, p.worker)
	stop := make(chan struct{})
	output := make(chan V)
	//defer func() {
	//	for range output {
	//		// TODO
	//	}
	//}()
	w := &writer[V]{
		ctx:  p.ctx,
		ch:   output,
		stop: stop,
	}
	// finish
	var o sync.Once
	done := func() {
		o.Do(func() {
			close(stop)
			close(output)
		})
	}
	var (
		once sync.Once
		e    atomic.Value
	)
	cancel := func(err error) {
		once.Do(func() {
			if err != nil {
				e.Store(err)
			} else {
				e.Store(errors.New("cancel func nil"))
			}
			clean(input)
			done()
		})
	}

	// merge
	go func() {
		defer func() {
			clean(coll)
			if r := recover(); r != nil {
				p.ec.write(r)
			}
			done()
		}()
		p.merger(coll, w, cancel)
	}()

	// split
	dt := &dtc[T, U]{
		input: input,
		coll:  coll,
	}
	ct := &ctrl{
		ctx:    p.ctx,
		stop:   stop,
		cancel: cancel,
	}
	go p.split(dt, ct)

	var (
		data V
		err  error
		ok   bool
	)
	select {
	case <-p.ctx.Done():
		cancel(context.DeadlineExceeded)
		err = context.DeadlineExceeded
	case <-p.ec.ch:
		clean(output)
		// TODO
	case data, ok = <-output:
		ee := e.Load()
		if ee != nil {
			err = ee.(error)
		} else if ok {

		} else {
			err = errors.New("output nil")
		}
	}
	return data, err
}

type dtc[T, U any] struct {
	input <-chan T
	coll  chan U
}

type ctrl struct {
	ctx    context.Context
	stop   <-chan struct{}
	cancel func(error)
}

type writer[T any] struct {
	ctx  context.Context
	stop <-chan struct{}
	ch   chan<- T
}

func (w *writer[T]) Write(v T) {
	select {
	case <-w.ctx.Done():
		return
	case <-w.stop:
		return
	default:
		w.ch <- v
	}
}

func (p *Parallel[T, U, V]) split(dtc *dtc[T, U], ctrl *ctrl) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(dtc.coll)
		clean(dtc.input)
	}()

	var stop atomic.Bool
	worker := make(chan struct{}, p.worker)
	w := &writer[U]{
		ctx:  p.ctx,
		ch:   dtc.coll,
		stop: ctrl.stop,
	}
	for !stop.Load() {
		select {
		case <-ctrl.ctx.Done():
			return
		case <-ctrl.stop:
			return
		default:
		}
		worker <- struct{}{}
		item, ok := <-dtc.input
		if !ok {
			<-worker
			return
		}
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					stop.Store(true)
					p.ec.write(r)
				}
				wg.Done()
				<-worker
			}()
			p.splitter(item, w, ctrl.cancel)
		}()
	}
}

func clean[T any](input <-chan T) {
	for range input {
	}
}
