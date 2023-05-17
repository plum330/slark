package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"net/http"
	"time"
)

type EngineConfig struct {
	Mode     string
	BasePath string
	http.FileSystem
}

func Engine(cfg *EngineConfig) ServerOption {
	return func(server *Server) {
		gin.SetMode(cfg.Mode)
		engine := server.Engine
		if cfg.FileSystem != nil {
			engine.StaticFS(fmt.Sprintf("%s/doc", cfg.BasePath), cfg.FileSystem)
		}
	}
}

func BuildRequestID(opts ...utils.Option) gin.HandlerFunc {
	cfg := &utils.Config{
		Builder: func() string {
			return utils.BuildRequestID()
		},
		RequestID: utils.RayID,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx *gin.Context) {
		rid := ctx.GetHeader(cfg.RequestID)
		if len(rid) == 0 {
			rid = cfg.Builder()
		}
		ctx.Header(cfg.RequestID, rid)
		ctx.Request = ctx.Request.WithContext(context.WithValue(context.Background(), cfg.RequestID, rid))
	}
}

func HandleMiddlewares(mw ...middleware.Middleware) gin.HandlerFunc {
	middle := middleware.ComposeMiddleware(mw...)
	return func(ctx *gin.Context) {
		reqCtx := ctx.Request.Context()
		_, err := middle(func(c context.Context, req interface{}) (interface{}, error) {
			ctx.Next()
			var err error
			status := ctx.Writer.Status()
			if status >= http.StatusBadRequest {
				err = errors.New(status, errors.UnknownReason, errors.UnknownReason)
			}
			return ctx.Writer, err
		})(reqCtx, ctx.Request)
		if err != nil {
			ctx.Abort()
			e := errors.FromError(err)
			if e.Message != errors.Panic {
				_ = ctx.Error(err)
			}
			rsp := &Response{
				Header: &Header{
					RayID: reqCtx.Value(utils.RayID),
				},
			}
			rsp.Code = int(e.Status.Code)
			rsp.Msg = e.Status.Message
			ctx.JSON(http.StatusOK, rsp)
		}
	}
}

func Log(l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		fields := map[string]interface{}{
			"request": fmt.Sprintf("%+v", ctx.Request),
			"start":   start.Format(time.RFC3339),
		}
		l.Log(ctx, logger.InfoLevel, fields, "request log")
		ctx.Next()
		for _, err := range ctx.Errors {
			ce, ok := err.Err.(*errors.Error)
			if !ok {
				fields = map[string]interface{}{
					"meta":    err.Meta,
					"type":    err.Type,
					"error":   fmt.Sprintf("%+v", err.Err),
					"latency": time.Since(start).Seconds(),
				}
				l.Log(ctx.Request.Context(), logger.ErrorLevel, fields, "unknown error")
			} else {
				fields = map[string]interface{}{
					"meta":    ce.Metadata,
					"reason":  ce.Reason,
					"code":    ce.Code,
					"surplus": ce.Surplus,
					"error":   fmt.Sprintf("%+v", err.Err),
					"latency": time.Since(start).Seconds(),
				}
				l.Log(ctx.Request.Context(), logger.ErrorLevel, fields, ce.Message)
			}
		}
		if len(ctx.Errors) == 0 {
			fields = map[string]interface{}{
				"latency":  time.Since(start).Seconds(),
				"response": ctx.Request.Response,
			}
			l.Log(ctx.Request.Context(), logger.InfoLevel, fields, "response log")
		}
	}
}
