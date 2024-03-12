package cache

import (
	"context"
	"github.com/go-slark/slark/logger"
	"sync/atomic"
	"time"
)

type Stat struct {
	key    string
	total  uint64
	hit    uint64
	miss   uint64
	dbFail uint64
}

func NewStat(key string) *Stat {
	stat := &Stat{
		key: key,
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		stat.loop(ticker)
	}()
	return stat
}

func (s *Stat) loop(ticker *time.Ticker) {
	for range ticker.C {
		total := atomic.SwapUint64(&s.total, 0)
		if total == 0 {
			continue
		}
		hit := atomic.SwapUint64(&s.hit, 0)
		ratio := 100 * float64(hit) / float64(total)
		miss := atomic.SwapUint64(&s.miss, 0)
		dbFail := atomic.SwapUint64(&s.dbFail, 0)
		fields := map[string]interface{}{
			"key":     s.key,
			"total":   total,
			"ratio":   ratio,
			"hit":     hit,
			"miss":    miss,
			"db_fail": dbFail,
		}
		logger.Log(context.TODO(), logger.DebugLevel, fields, "cache stat")
	}
}

func (s *Stat) IncrementTotal() {
	atomic.AddUint64(&s.total, 1)
}

func (s *Stat) IncrementHit() {
	atomic.AddUint64(&s.hit, 1)
}

func (s *Stat) IncrementMiss() {
	atomic.AddUint64(&s.miss, 1)
}

func (s *Stat) IncrementDbFails() {
	atomic.AddUint64(&s.dbFail, 1)
}
