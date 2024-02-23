package bulk

import (
	"sync"
	"time"
)

type Batch struct {
	task  *Task
	batch *batchTasker
}

func NewBatch(f func([]any), opts ...BatchOption) *Batch {
	batch := &batchTasker{
		l:        sync.Mutex{},
		max:      1000,
		f:        f,
		interval: time.Second,
	}
	for _, opt := range opts {
		opt(batch)
	}
	batch.tasks = make([]any, 0, batch.max)
	bt := &Batch{
		task:  NewTask(batch, batch.interval),
		batch: batch,
	}
	return bt
}

func (b *Batch) Submit(v any) {
	b.task.Submit(v)
}

func (b *Batch) Force() {
	b.task.Force()
}

type batchTasker struct {
	tasks    []any
	l        sync.Mutex
	max      int
	f        func([]any)
	interval time.Duration
}

type BatchOption func(*batchTasker)

func BatchMax(max int) BatchOption {
	return func(b *batchTasker) {
		b.max = max
	}
}

func BatchInterval(interval time.Duration) BatchOption {
	return func(b *batchTasker) {
		b.interval = interval
	}
}

func (b *batchTasker) Submit(v any) bool {
	b.l.Lock()
	defer b.l.Unlock()
	b.tasks = append(b.tasks, v)
	return len(b.tasks) >= b.max
}

func (b *batchTasker) Fetch() []any {
	b.l.Lock()
	defer b.l.Unlock()
	task := b.tasks
	b.tasks = make([]any, 0, b.max)
	return task
}

func (b *batchTasker) Do(tasks []any) {
	b.f(tasks)
}
