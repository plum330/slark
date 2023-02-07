package retry

import (
	"fmt"
	"github.com/go-slark/slark/errors"
	"testing"
	"time"
)

// 200ms 400ms 800ms 1.6s
func TestBackoffRetry(t *testing.T) {
	opt := NewOption(Debug(true))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}

// 200ms 400ms 800ms 1s
func TestBackoffRetryWithMaxDelay(t *testing.T) {
	opt := NewOption(Debug(true), MaxDelay(1*time.Second))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}

// 100ms 100ms 100ms 100ms
func TestFixed(t *testing.T) {
	opt := NewOption(Debug(true), Function(Fixed))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}

func TestRandom(t *testing.T) {
	opt := NewOption(Debug(true), Function(Random))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}

// 300ms 500ms 900ms 1.7s
func TestGroup(t *testing.T) {
	opt := NewOption(Debug(true), Function(Group(Fixed, BackOff)))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}

// 300ms 500ms 900ms 1s
func TestGroupWithMaxDelay(t *testing.T) {
	opt := NewOption(Debug(true), MaxDelay(1*time.Second), Function(Group(Fixed, BackOff)))
	err := opt.Retry(func() error {
		return errors.NewError(600, "test", "test")
	})
	fmt.Println(err)
}
