package future

import (
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/mr"
	"testing"
)

func TestParallel(t *testing.T) {
	data := []int{100, 200, 700}
	producer := func(input chan<- int) {
		for _, d := range data {
			input <- d
		}
	}
	splitter := func(item int, w mr.Writer[int], cancel func(error)) {
		// 填入item
		w.Write(item)
	}
	merger := func(ch <-chan int, w mr.Writer[int], cancel func(error)) {
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
	splitter := func(item int, w mr.Writer[int], cancel func(error)) {
		cancel(nil)
		//cancel(errors.New("split cancel"))
		// 填入item
		w.Write(item)
	}
	merger := func(ch <-chan int, w mr.Writer[int], cancel func(error)) {
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

func TestExec(t *testing.T) {
	err := Exec(func() error {
		fmt.Println("11111")
		return nil
	}, func() error {
		fmt.Println("22222")
		return errors.New("error")
	}, func() error {
		fmt.Println("777777")
		return errors.New("error 333")
	})
	fmt.Println("error:", err)
}

func TestVoidExec(t *testing.T) {
	VoidExec(func() {
		fmt.Println("33333")
	}, func() {
		fmt.Println("55555")
	})
}
