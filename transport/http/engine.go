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
		ctx.Next()
		for _, err := range ctx.Errors {
			ce, ok := err.Err.(*errors.Error)
			if !ok {
				fields := map[string]interface{}{
					"meta":  err.Meta,
					"type":  err.Type,
					"error": fmt.Sprintf("%+v", err.Err),
				}
				l.Log(ctx.Request.Context(), logger.ErrorLevel, fields, "unknown error")
			} else {
				fields := map[string]interface{}{
					"meta":    ce.Metadata,
					"reason":  ce.Reason,
					"code":    ce.Code,
					"surplus": ce.Surplus,
					"error":   fmt.Sprintf("%+v", err.Err),
				}
				l.Log(ctx.Request.Context(), logger.ErrorLevel, fields, ce.Message)
			}
		}
	}
}
