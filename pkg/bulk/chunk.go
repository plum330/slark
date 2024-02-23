package bulk

import (
	"sync"
	"time"
)

type Chunk struct {
	task  *Task
	chunk *chunkTasker
}

func NewChunk(f func([]any), opts ...ChunkOption) *Chunk {
	ct := &chunkTasker{
		tasks:    make([]any, 0),
		l:        sync.Mutex{},
		f:        f,
		max:      1024 * 1024,
		interval: time.Second,
	}
	for _, opt := range opts {
		opt(ct)
	}
	ck := &Chunk{
		task:  NewTask(ct, ct.interval),
		chunk: ct,
	}
	return ck
}

func (c *Chunk) Submit(v any, size int) {
	c.task.Submit(chunk{
		task: v,
		size: size,
	})
}

func (c *Chunk) Force() {
	c.task.Force()
}

type chunkTasker struct {
	tasks     []any
	l         sync.Mutex
	f         func([]any)
	max, size int
	interval  time.Duration
}

type ChunkOption func(*chunkTasker)

func ChunkMax(max int) ChunkOption {
	return func(c *chunkTasker) {
		c.max = max
	}
}

func ChunkInterval(interval time.Duration) ChunkOption {
	return func(c *chunkTasker) {
		c.interval = interval
	}
}

type chunk struct {
	task any
	size int
}

func (c *chunkTasker) Submit(v any) bool {
	ck, _ := v.(chunk)
	c.l.Lock()
	defer c.l.Unlock()
	c.tasks = append(c.tasks, ck.task)
	c.size += ck.size
	return c.size > c.max
}

func (c *chunkTasker) Fetch() []any {
	c.l.Lock()
	defer c.l.Unlock()
	tasks := c.tasks
	c.tasks = make([]any, 0)
	c.size = 0
	return tasks
}

func (c *chunkTasker) Do(v []any) {
	c.f(v)
}
