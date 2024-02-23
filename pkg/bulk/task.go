package bulk

import (
	"context"
	"github.com/go-slark/slark/pkg/routine"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

/*
	批量提交任务
	缓冲一部分任务，惰性提交
	延迟任务提交
*/

// 任务方需保证任务池线程安全

type Tasker interface {
	Submit(any) bool
	Fetch() []any
	Do([]any)
}

type Task struct {
	tasker   Tasker
	interval time.Duration
	ticker   func(tm time.Duration) *time.Ticker
	tasks    atomic.Int64 // 任务量
	routine  atomic.Bool
	cmd      chan []any
}

func NewTask(tasker Tasker, interval time.Duration) *Task {
	task := &Task{
		tasker:   tasker,
		interval: interval,
		ticker: func(tm time.Duration) *time.Ticker {
			return time.NewTicker(tm)
		},
		tasks:   atomic.Int64{},
		routine: atomic.Bool{},
		cmd:     make(chan []any, 1),
	}
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
		<-signals
		task.Force()
	}()
	return task
}

func (t *Task) Submit(v any) {
	tasks := t.submit(v)
	if len(tasks) > 0 {
		t.cmd <- tasks
	}
}

func (t *Task) Force() {
	tasks := t.tasker.Fetch()
	t.do(tasks)
}

func (t *Task) submit(v any) []any {
	// 提交任务达到max,触发任务执行
	if t.tasker.Submit(v) {
		t.tasks.Add(1)
		return t.tasker.Fetch()
	}

	if !t.routine.Load() {
		t.flush()
		t.routine.Store(true)
	}
	return nil
}

func (t *Task) flush() {
	go func() {
		ticker := t.ticker(t.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Force()
				// TODO
			case tasks := <-t.cmd:
				t.tasks.Add(-1)
				t.do(tasks)
			}
		}
	}()
}

func (t *Task) do(tasks []any) {
	if len(tasks) == 0 {
		return
	}
	routine.Go(context.TODO(), func() {
		t.tasker.Do(tasks)
	})
}
