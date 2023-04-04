package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/logger/engine_logger/mw_logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type EngineParam struct {
	Mode        string
	BaseUrl     string
	Routers     []func(r gin.IRouter)
	HandlerFunc []gin.HandlerFunc
	http.FileSystem
	logger.Logger
}

func Engine(param *EngineParam) ServerOption {
	return func(server *Server) {
		gin.SetMode(param.Mode)
		engine := server.Engine
		engine.Use(BuildRequestId())
		engine.Use(mw_logger.ErrLogger(param.Logger))
		if param.FileSystem != nil {
			engine.StaticFS(fmt.Sprintf("%s/doc", param.BaseUrl), param.FileSystem)
		}
		engine.Use(param.HandlerFunc...)
		g := engine.Group(param.BaseUrl)
		for _, router := range param.Routers {
			router(g)
		}
	}
}

func BuildRequestId(opts ...pkg.Option) gin.HandlerFunc {
	cfg := &pkg.Config{
		Builder: func() string {
			return pkg.BuildRequestID()
		},
		RequestId: pkg.TraceID,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx *gin.Context) {
		rid := ctx.GetHeader(cfg.RequestId)
		if len(rid) == 0 {
			rid = cfg.Builder()
		}
		ctx.Header(cfg.RequestId, rid)
		ctx.Request = ctx.Request.WithContext(context.WithValue(context.Background(), cfg.RequestId, rid))
	}
}

func GetRequestId(ctx *gin.Context) string {
	return ctx.Writer.Header().Get(pkg.TraceID)
}

type Header struct {
	Code    int         `json:"code"`
	TraceID interface{} `json:"trace_id"`
	Msg     string      `json:"msg"`
}

type Response struct {
	*Header
	proto.Message
}

func (r Response) Render(w http.ResponseWriter) (err error) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
	codec := encoding.GetCodec("json")
	hb, err := codec.Marshal(r.Header)
	if err != nil {
		return err
	}
	pb, err := codec.Marshal(r.Message)
	if err != nil {
		return err
	}
	data := make([]byte, 0, len(hb)+len(pb)+8)
	data = append(data, hb[:len(hb)-1]...)
	data = append(data, []byte(`,"data":`)...)
	data = append(data, pb...)
	data = append(data, '}')
	_, err = w.Write(data)
	return err
}

func (r Response) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

func Result(out proto.Message, err error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rsp := &Response{
			Header: &Header{
				TraceID: ctx.Request.Context().Value(pkg.TraceID),
			},
		}
		rsp.Msg = "成功"
		rsp.Message = out
		if err != nil {
			e := errors.ParseErr(err)
			rsp.Code = int(e.Status.Code)
			rsp.Msg = e.Status.Message
			_ = ctx.Error(e)
			ctx.Abort()
		}
		ctx.JSON(http.StatusOK, rsp)
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
			e := errors.ParseErr(err)
			if e.Message != errors.Panic {
				_ = ctx.Error(err)
			}
			rsp := &Response{
				Header: &Header{
					TraceID: reqCtx.Value(pkg.TraceID),
				},
			}
			rsp.Code = int(e.Status.Code)
			rsp.Msg = e.Status.Message
			ctx.JSON(http.StatusOK, rsp)
		}
	}
}
