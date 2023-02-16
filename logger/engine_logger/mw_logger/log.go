package mw_logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/logger/engine_logger/http_logger"
)

func ErrLogger(l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		context := ctx.Request.Context()
		for _, err := range ctx.Errors {
			ce, ok := err.Err.(*errors.Error)
			if !ok {
				fields := map[string]interface{}{
					"meta":  err.Meta,
					"type":  err.Type,
					"error": fmt.Sprintf("%+v", err.Err),
				}
				l.Log(context, logger.ErrorLevel, fields, "系统异常")
			} else {
				fields := map[string]interface{}{
					"meta":    ce.Metadata,
					"reason":  ce.Reason,
					"code":    ce.Code,
					"surplus": ce.Surplus,
					"error":   fmt.Sprintf("%+v", err.Err),
				}
				l.Log(context, logger.ErrorLevel, fields, ce.Message)
			}
		}
	}
}

func Logger(logger logger.Logger, excludePaths ...string) gin.HandlerFunc {
	l := http_logger.AccessLoggerConfig{
		Logger:         logger,
		BodyLogPolicy:  http_logger.LogAllBodies,
		MaxBodyLogSize: 1024 * 16, //16k
		DropSize:       1024 * 10, //10k
	}

	l.ExcludePaths = map[string]struct{}{}
	for _, excludePath := range excludePaths {
		l.ExcludePaths[excludePath] = struct{}{}
	}
	return http_logger.New(l)
}
