package limiter

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-slark/slark/logger"
	"golang.org/x/time/rate"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TBLimiter struct {
	rate     int
	burst    int
	redis    *redis.Client
	alive    uint32
	l        sync.Mutex
	tokenKey string
	tsKey    string
	monitor  bool
	limiter  *rate.Limiter // local limiter
	logger   logger.Logger
}

func NewTBLimiter(limit, burst int, redis *redis.Client, key string) *TBLimiter {
	return &TBLimiter{
		rate:     limit,
		burst:    burst,
		redis:    redis,
		tokenKey: fmt.Sprintf("token_bucket:key:%s:tokens", key),
		tsKey:    fmt.Sprintf("token_bucket:key:%s:ts", key),
		alive:    1,
		limiter:  rate.NewLimiter(rate.Every(time.Second/time.Duration(limit)), burst),
		logger:   logger.GetLogger(),
	}
}

func (l *TBLimiter) AllowN(now time.Time, n int) bool {
	return l.reserveN(context.Background(), now, n)
}

const script string = `
	local rate = tonumber(ARGV[1])
	local capacity = tonumber(ARGV[2])
	local now_ts = tonumber(ARGV[3])
	local request_tokens = tonumber(ARGV[4])
	local fill_time = capacity/rate
	local ttl = math.floor(fill_time*2)
	local last_tokens = tonumber(redis.call("GET", KEYS[1]))
	if last_tokens == nil then
    	last_tokens = capacity
	end

	local last_refresh_ts = tonumber(redis.call("GET", KEYS[2]))
	if last_refresh_ts == nil then
    	last_refresh_ts = 0
	end

	local delta = math.max(0, now_ts-last_refresh_ts)
	local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
	local allowed = filled_tokens >= request_tokens
	local new_tokens = filled_tokens
	if allowed then
    	new_tokens = filled_tokens - request_tokens
	end

	redis.call("SETEX", KEYS[1], ttl, new_tokens)
	redis.call("SETEX", KEYS[2], ttl, now_ts)

	return allowed
`

func (l *TBLimiter) reserveN(ctx context.Context, now time.Time, n int) bool {
	if atomic.LoadUint32(&l.alive) == 0 {
		return l.limiter.AllowN(now, n)
	}

	keys := []string{l.tokenKey, l.tsKey}
	args := []string{
		strconv.Itoa(l.rate),
		strconv.Itoa(l.burst),
		strconv.FormatInt(now.Unix(), 10),
		strconv.Itoa(n),
	}
	result, err := redis.NewScript(script).Run(l.redis, keys, args).Result()
	if err == redis.Nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		l.logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": fmt.Sprintf("%+v", err)}, "limit fail")
		return false
	}
	if err != nil {
		l.logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": fmt.Sprintf("%+v", err)}, "script run fail, use local limiter")
		l.observe()
		return l.limiter.AllowN(now, n)
	}

	code, ok := result.(int64)
	if !ok {
		l.logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{"result": fmt.Sprintf("%+v", result)}, "script result invalid, use local limiter")
		l.observe()
		return l.limiter.AllowN(now, n)
	}
	return code == 1
}

func (l *TBLimiter) observe() {
	l.l.Lock()
	defer l.l.Unlock()

	if l.monitor {
		return
	}

	l.monitor = true
	atomic.StoreUint32(&l.alive, 0)

	go l.health()
}

func (l *TBLimiter) health() {
	tk := time.NewTicker(100 * time.Millisecond)
	defer func() {
		tk.Stop()
		l.l.Lock()
		l.monitor = false
		l.l.Unlock()
	}()

	for range tk.C {
		var ping bool
		result, err := l.redis.Ping().Result()
		if err != nil {
			ping = false
		} else {
			ping = result == "PONG"
		}
		if ping {
			atomic.StoreUint32(&l.alive, 1)
			return
		}
	}
}
