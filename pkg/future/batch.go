package future

import (
	"context"
	"errors"
	"sync"
	"time"
)

type BatchOption func(*Batch)

func WithInterval(interval time.Duration) BatchOption {
	return func(o *Batch) {
		o.interval = interval
	}
}

func WithNum(num int) BatchOption {
	return func(o *Batch) {
		o.num = num
	}
}

func WithSize(size int) BatchOption {
	return func(o *Batch) {
		o.size = size
	}
}

func WithWorker(worker int) BatchOption {
	return func(o *Batch) {
		o.worker = worker
	}
}

func WithExec(exec func(context.Context, map[string][]any)) BatchOption {
	return func(o *Batch) {
		o.exec = exec
	}
}

func WithSharding(sharding func(string) int) BatchOption {
	return func(o *Batch) {
		o.sharding = sharding
	}
}

type data struct {
	key   string
	value any
}

type Batch struct {
	exec              func(context.Context, map[string][]any)
	sharding          func(string) int
	chs               []chan *data
	wg                sync.WaitGroup
	interval          time.Duration
	num, size, worker int
}

func NewBatch(opts ...BatchOption) *Batch {
	batch := &Batch{
		interval: time.Second,
		num:      100,
		size:     100,
		worker:   5,
	}
	for _, opt := range opts {
		opt(batch)
	}
	batch.chs = make([]chan *data, batch.worker)
	for i := 0; i < batch.worker; i++ {
		batch.chs[i] = make(chan *data, batch.size)
	}
	return batch
}

func (b *Batch) Run() error {
	if b.exec == nil || b.sharding == nil {
		return errors.New("exec and sharding func not exists")
	}
	b.wg.Add(len(b.chs))
	for index, ch := range b.chs {
		go b.do(index, ch)
	}
	return nil
}

func (b *Batch) do(index int, ch <-chan *data) {
	var (
		num          int
		stop, adjust bool
	)
	msg := make(map[string][]any)
	interval := b.interval
	if index == 0 {
		// 调整时间间隔，goroutine lb
		interval = time.Duration(int64(b.interval) * int64(index) / int64(b.worker))
		adjust = true
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case item := <-ch:
			if item == nil {
				stop = true
				break
			}
			num++
			msg[item.key] = append(msg[item.key], item.value)
			if num < b.num {
				continue
			}
		case <-ticker.C:
			if adjust {
				ticker.Stop()
				ticker = time.NewTicker(b.interval)
				adjust = false
			}
		}
		if len(msg) > 0 {
			b.exec(context.TODO(), msg)
			msg = make(map[string][]any)
			num = 0
		}
		if stop {
			ticker.Stop()
			b.wg.Done()
			return
		}
	}
}

func (b *Batch) Add(key string, value any) error {
	select {
	case b.chs[b.sharding(key)%b.worker] <- &data{
		key:   key,
		value: value,
	}:
	default:
		return errors.New("data channel full")
	}
	return nil
}

func (b *Batch) Stop() {
	for _, ch := range b.chs {
		ch <- nil
	}
	b.wg.Wait()
}
