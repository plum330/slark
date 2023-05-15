package main

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
{{$svrType := .ServiceType}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(srv {{.ServiceType}}HTTPServer) {
	r := http.NewRouter()
	{{- range .Methods}}
	r.Handle("{{.Method}}", "{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) http.HandlerFunc {
	return func(ctx *http.Context) error {
		var (
			in {{.Request}}
			out *{{.Reply}}
			err error
			newCtx context.Context
		)
		
		{{- if .HasBody}}
		err = ctx.BindBody(&in{{.Body}})
		if err != nil {
			return err
		}

		{{- if .not (eq .Body "")}}
		err = ctx.BindQuery(&in)
		if err != nil {
			return err
		}
		{{- end}}
		
		{{- else}}
		err = ctx.BindQuery(&in)
		if err != nil {
			return err
		}

		err = ctx.BindVars(&in)
		if err != nil {
			return err
		}
		{{- end}}

		err = in.ValidateAll()
		if err != nil {
			return err
		}

		newCtx = context.WithValue(ctx.Request.Context(), utils.Token, ctx.GetHeader(utils.Token))
		out, err = srv.{{.Name}}(newCtx, &in)
		if err != nil {
			return err
		}
		return ctx.Result(0, out.(*{{.Reply}}){{.ResponseBody}})
	}
}
{{end}}
`

type serviceDesc struct {
	PackageName string
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	// http_rule
	Path         string
	Method       string
	HasVars      bool
	HasBody      bool
	Body         string
	ResponseBody string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
