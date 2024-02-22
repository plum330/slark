//go:build linux

package pprof

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
	"syscall"
	"time"
)

func init() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGUSR1, syscall.SIGUSR2)
		for {
			switch <-signals {
			case syscall.SIGUSR1: // kill -usr1 pid
			case syscall.SIGUSR2: // kill -usr2 pid
				pf := profile{make([]func(), 0, 6)}
				pf.start()
				time.AfterFunc(time.Minute, pf.stop)
			}
		}
	}()
}

func file(name string) string {
	cmd := path.Base(os.Args[0])
	pid := syscall.Getpid()
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%d-%s-%s.pprof", cmd, pid, name, time.Now().Format("2006-01-02 15:04:05.000")))
}

type profile struct {
	closers []func()
}

func (p *profile) block() {
	fn := file("block")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create block file error")
		return
	}
	runtime.SetBlockProfileRate(1)
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "block file start")
	p.closers = append(p.closers, func() {
		_ = pprof.Lookup("block").WriteTo(f, 0)
		_ = f.Close()
		runtime.SetBlockProfileRate(0)
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "block file stop")
	})
}

func (p *profile) cpu() {
	fn := file("cpu")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create cpu file error")
		return
	}
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "cpu file start")
	_ = pprof.StartCPUProfile(f)
	p.closers = append(p.closers, func() {
		pprof.StopCPUProfile()
		_ = f.Close()
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "cpu file stop")
	})
}

func (p *profile) memory() {
	fn := file("mem")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create memory file error")
		return
	}
	old := runtime.MemProfileRate
	runtime.MemProfileRate = 4096
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "memory file start")
	p.closers = append(p.closers, func() {
		_ = pprof.Lookup("heap").WriteTo(f, 0)
		_ = f.Close()
		runtime.MemProfileRate = old
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "memory file stop")
	})
}

func (p *profile) mutex() {
	fn := file("mutex")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create mutex file error")
		return
	}
	runtime.SetMutexProfileFraction(1)
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "mutex file start")
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("mutex"); mp != nil {
			_ = mp.WriteTo(f, 0)
		}
		_ = f.Close()
		runtime.SetMutexProfileFraction(0)
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "mutex file stop")
	})
}

func (p *profile) threadCreate() {
	fn := file("threadcreate")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create thread create file error")
		return
	}
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "thread create file start")
	p.closers = append(p.closers, func() {
		if mp := pprof.Lookup("threadcreate"); mp != nil {
			_ = mp.WriteTo(f, 0)
		}
		_ = f.Close()
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "thread create file stop")
	})
}

func (p *profile) trace() {
	fn := file("trace")
	f, err := os.Create(fn)
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "create trace file error")
		return
	}
	if err = trace.Start(f); err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "trace start error")
		return
	}
	logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "trace file start")
	p.closers = append(p.closers, func() {
		trace.Stop()
		logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"file": fn}, "trace file start")
	})
}

var started int32

func (p *profile) start() {
	// 采样分析正在进行
	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return
	}

	p.cpu()
	p.memory()
	p.mutex()
	p.block()
	p.trace()
	p.threadCreate()

	//go func() {
	//	c := make(chan os.Signal, 1)
	//	signal.Notify(c, syscall.SIGINT)
	//	<-c
	//	p.stop()
	//	signal.Reset()
	//	syscall.Kill(os.Getpid(), syscall.SIGINT)
	//}()
}

func (p *profile) stop() {
	atomic.StoreInt32(&started, 0)
	for _, closer := range p.closers {
		closer()
	}
}
