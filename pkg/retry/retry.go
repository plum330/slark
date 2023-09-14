package retry

/*
	最简单的重试是循环执行相关代码，存在的问题：
	1.重试之间没有时间间隔，如网络原因造成请求失败，若重试请求间隔时间太短，这种重试无意义
	2.发生错误时，不能根据错误类型调整重试策略
	3.可能造成惊群问题 （Thundering Herd Problem），当服务端一次断开大量连接，客户端会同时发送重试请求，极易造成惊群问题
*/

import (
	"context"
	"github.com/go-slark/slark/logger"
	"math"
	"math/rand"
	"time"
)

type Func func(int, *Option) time.Duration

type Option struct {
	retry     int
	backoff   int
	delay     time.Duration
	maxDelay  time.Duration
	maxJitter time.Duration
	f         Func
	timer     func(time.Duration) <-chan time.Time
	ctx       context.Context
	debug     bool
}

func NewOption(opts ...Opt) *Option {
	o := &Option{
		retry:     5,
		delay:     100 * time.Millisecond,
		maxJitter: 100 * time.Millisecond,
		f:         BackOff,
		ctx:       context.TODO(),
		timer: func(d time.Duration) <-chan time.Time {
			timer := time.NewTimer(d)
			defer timer.Stop()
			c := make(chan time.Time, 1)
			select {
			case tm := <-timer.C:
				c <- tm
			}
			return c
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Opt func(*Option)

func Retry(retry int) Opt {
	return func(o *Option) {
		if retry > 0 {
			o.retry = retry
		}
	}
}

func Delay(delay time.Duration) Opt {
	return func(o *Option) {
		o.delay = delay
	}
}

func MaxDelay(maxDelay time.Duration) Opt {
	return func(o *Option) {
		o.maxDelay = maxDelay
	}
}

func MaxJitter(maxJitter time.Duration) Opt {
	return func(o *Option) {
		o.maxJitter = maxJitter
	}
}

func Function(f Func) Opt {
	return func(o *Option) {
		o.f = f
	}
}

func Context(ctx context.Context) Opt {
	return func(o *Option) {
		o.ctx = ctx
	}
}

func Debug(debug bool) Opt {
	return func(o *Option) {
		o.debug = debug
	}
}

func BackOff(n int, o *Option) time.Duration {
	// 1 << 63 would overflow signed int64 (time.Duration), thus 62.
	max := 62
	if o.backoff == 0 {
		if o.delay <= 0 {
			o.delay = 1
		}
		o.backoff = max - int(math.Floor(math.Log2(float64(o.delay))))
	}
	if n > o.backoff {
		n = o.backoff
	}
	return o.delay << n
}

func Fixed(_ int, o *Option) time.Duration {
	return o.delay
}

func Random(_ int, o *Option) time.Duration {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Duration(rand.Int63n(int64(o.maxJitter)))
}

func Group(fs ...Func) Func {
	return func(n int, o *Option) time.Duration {
		var total time.Duration
		for _, f := range fs {
			total += f(n, o)
			if total > math.MaxInt64 {
				total = math.MaxInt64
			}
		}
		return total
	}
}

func Timer(timer func(d time.Duration) <-chan time.Time) Opt {
	return func(o *Option) {
		o.timer = timer
	}
}

func (o *Option) Retry(fn func() error) error {
	var err error
	for n := 1; n <= o.retry; n++ {
		err = fn()
		if err == nil {
			break
		}

		if n == o.retry {
			break
		}

		select {
		case <-o.timer(delay(o, n)):
			// TODO
		}
	}
	return err
}

func delay(o *Option, n int) time.Duration {
	delayTime := o.f(n, o)
	if o.maxDelay > 0 && delayTime > o.maxDelay {
		delayTime = o.maxDelay
	}
	if o.debug {
		logger.Log(context.TODO(), logger.DebugLevel, map[string]interface{}{"times": n, "delay_time": delayTime, "max_delay": o.maxDelay}, "正在进行重试")
	}
	return delayTime
}
