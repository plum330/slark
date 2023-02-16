package logging

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/errors"
	"testing"
	"time"
)

type mockLogger struct{}

func (l *mockLogger) Log(ctx context.Context, level uint, fields map[string]interface{}, v ...interface{}) {
	fmt.Printf("fields:%+v\n", fields)
}

func TestLoggerError(t *testing.T) {
	_, _ = Log(&mockLogger{})(func(ctx context.Context, req interface{}) (interface{}, error) {
		fmt.Println("test logger error")
		return nil, errors.BadRequest("bad request", "bad request")
	})(context.TODO(), 3)
}

func TestLogger(t *testing.T) {
	time.Sleep(time.Second)
	_, _ = Log(&mockLogger{})(func(ctx context.Context, req interface{}) (interface{}, error) {
		fmt.Println("test logger")
		return nil, nil
	})(context.TODO(), 1)
}
