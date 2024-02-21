package future

import (
	"errors"
	"testing"
)

func TestParallel(t *testing.T) {
	data := []int{100, 200, 700}
	producer := func(input chan<- int) {
		for _, d := range data {
			input <- d
		}
	}
	splitter := func(item int, w Writer[int], cancel func(error)) {
		// 填入item
		w.Write(item)
	}
	merger := func(ch chan int, w Writer[int], cancel func(error)) {
		var sum int
		for item := range ch {
			sum += item
		}
		w.Write(sum)
	}
	p := NewParallel(
		Producer[int, int, int](producer),
		Splitter[int, int, int](splitter),
		Merger[int, int, int](merger),
	)
	v, err := p.Do()
	if err != nil {
		t.Errorf("error:%+v", err)
		return
	}
	t.Logf("result:%v", v)
}

func TestParallelWithCancel(t *testing.T) {
	data := []int{100, 200, 700}
	producer := func(input chan<- int) {
		for _, d := range data {
			input <- d
		}
	}
	splitter := func(item int, w Writer[int], cancel func(error)) {
		cancel(nil)
		//cancel(errors.New("split cancel"))
		// 填入item
		w.Write(item)
	}
	merger := func(ch chan int, w Writer[int], cancel func(error)) {
		cancel(errors.New("merge cancel"))
		var sum int
		for item := range ch {
			sum += item
		}
		w.Write(sum)
	}
	p := NewParallel(
		Producer[int, int, int](producer),
		Splitter[int, int, int](splitter),
		Merger[int, int, int](merger),
	)
	v, err := p.Do()
	if err != nil {
		t.Errorf("error:%+v", err)
		return
	}
	t.Logf("result:%v", v)
}
