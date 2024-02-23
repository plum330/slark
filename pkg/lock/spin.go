package lock

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	lock uint32
}

func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapUint32(&l.lock, 0, 1) {
		runtime.Gosched()
	}
}

func (l *SpinLock) Unlock() {
	atomic.StoreUint32(&l.lock, 0)
}
