package mw_logger

import (
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
			ce, ok := err.Err.(*errors.CustomError)
			if !ok {
				l.Log(context, logger.ErrorLevel, map[string]interface{}{"meta": err.Meta, "error": err.Err}, "系统异常")
			} else {
				fields := map[string]interface{}{
					"surplus": ce.Surplus,
					"meta":    ce.Metadata,
					"code":    ce.Code,
					"error":   ce.GetError(),
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
