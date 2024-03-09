package limit

import (
	"golang.org/x/time/rate"
	"time"
)

type RateLimit struct {
	rate, burst int
	limiter     *rate.Limiter
}

func NewRateLimiter() *RateLimit {
	return &RateLimit{
		rate:    0,
		burst:   0,
		limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(1)), 1),
	}
}

func (r *RateLimit) Pass() error {
	r.limiter.AllowN(time.Now(), 100)
	return nil
}
