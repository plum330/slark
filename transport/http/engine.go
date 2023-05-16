package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"net/http"
)

type EngineConfig struct {
	Mode    string
	BaseUrl string
	http.FileSystem
}

func Engine(param *EngineConfig) ServerOption {
	return func(server *Server) {
		gin.SetMode(param.Mode)
		engine := server.Engine
		if param.FileSystem != nil {
			engine.StaticFS(fmt.Sprintf("%s/doc", param.BaseUrl), param.FileSystem)
		}
		engine.Group(param.BaseUrl)
	}
}

func BuildRequestID(opts ...utils.Option) middleware.HTTPMiddleware {
	cfg := &utils.Config{
		Builder: func() string {
			return utils.BuildRequestID()
		},
		RequestID: utils.RayID,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(cfg.RequestID)
			if len(rid) == 0 {
				rid = cfg.Builder()
			}
			w.Header().Set(cfg.RequestID, rid)
			r = r.WithContext(context.WithValue(context.Background(), cfg.RequestID, rid))
		})
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
