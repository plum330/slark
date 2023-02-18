package http

import (
	"context"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/logger/engine_logger/mw_logger"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type EngineParam struct {
	Mode         string
	BaseUrl      string
	AccessLog    bool
	Pprof        bool
	ExcludePaths []string
	Routers      []func(r gin.IRouter)
	HandlerFunc  []gin.HandlerFunc
	http.FileSystem
	logger.Logger
}

func Engine(param *EngineParam) ServerOption {
	return func(server *Server) {
		gin.SetMode(param.Mode)
		engine := server.Engine
		engine.Use(BuildRequestId())
		engine.Use(mw_logger.ErrLogger(param.Logger))
		if param.AccessLog {
			engine.Use(mw_logger.Logger(param.Logger, param.ExcludePaths...))
		}
		if param.Pprof {
			pprof.Register(engine)
		}
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

type ProtoJson struct {
	Code    int
	TraceID interface{}
	Msg     string
	Data    proto.Message
}

var MarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: true,
}

func (p ProtoJson) Render(w http.ResponseWriter) (err error) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
	jsonBytes, err := MarshalOptions.Marshal(p.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		panic(err)
	}
	return
}

func (p ProtoJson) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
}

func Result(out proto.Message, err error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rsp := &ProtoJson{TraceID: ctx.Request.Context().Value(pkg.TraceID)}
		rsp.Msg = "成功"
		rsp.Data = out
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
			if status != http.StatusOK {
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
			rsp := &ProtoJson{TraceID: reqCtx.Value(pkg.TraceID)}
			rsp.Code = int(e.Status.Code)
			rsp.Msg = e.Status.Message
			ctx.JSON(http.StatusOK, rsp)
		}
	}
}
