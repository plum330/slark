package shedding

import (
	"sync/atomic"
	"time"
)

// cpu max_conn overload

type Shedding struct {
	name   string
	time   time.Duration
	total  atomic.Int64
	pass   atomic.Int64
	reject atomic.Int64
}

func NewShedding(name string) *Shedding {
	s := &Shedding{
		name:   name,
		time:   time.Minute,
		total:  atomic.Int64{},
		pass:   atomic.Int64{},
		reject: atomic.Int64{},
	}
	go s.run()
	return s
}

func (s *Shedding) reset() {

}

func (s *Shedding) run() {
	tk := time.NewTicker(s.time)
	defer tk.Stop()
	for range tk.C {
		s.reset()
	}
}

func (s *Shedding) IncrTotal() {
	s.total.Add(1)
}

func (s *Shedding) IncrPass() {
	s.pass.Add(1)
}

func (s *Shedding) IncrReject() {
	s.reject.Add(1)
}
