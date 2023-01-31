package middleware

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestRecovery(t *testing.T) {
	_, _ = HandleMiddleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("6666")
		return nil, nil
	}, Recovery(logger.NewLog(logger.WithFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		PrettyPrint:     false,
	}))))(context.TODO(), "$$$")
}
