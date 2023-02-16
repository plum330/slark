package recovery

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"testing"
)

func TestRecovery(t *testing.T) {
	_, _ = middleware.ComposeMiddleware(Recovery(logger.NewLog()))(func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("6666")
		return nil, nil
	})(context.TODO(), "$$$")
}
