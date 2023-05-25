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

func Register{{.ServiceType}}HTTPServer(s *http.Server, srv {{.ServiceType}}HTTPServer) {
	r := http.NewRouter(s)
	{{- range .Methods}}
	r.Handle("{{.Method}}", "{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) http.HandlerFunc {
	return func(ctx *http.Context) error {
		var (
			in {{.Request}}
			out interface{}
			err error
		)
		
		{{- if .HasBody}}
		err = ctx.ShouldBind(&in)
		if err != nil {
			return errors.BadRequest(errors.FormatError, err.Error())
		}

		{{- else if .HasQuery}}
		err = ctx.ShouldBindQuery(&in)
		if err != nil {
			return errors.BadRequest(errors.FormatError, err.Error())
		}

		{{- else}}
		err = ctx.ShouldBindURI(&in)
		if err != nil {
			return errors.BadRequest(errors.FormatError, err.Error())	
		}
		{{- end}}

		out, err = ctx.Handle(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req)
		})(ctx.Context(), &in)
		if err != nil {
			return err
		}
		return ctx.Result(out.(*{{.Reply}}))
	}
}
{{end}}
`

type serviceDesc struct {
	PackageName string
	ServiceType string
	ServiceName string
	Metadata    string
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
	HasQuery     bool
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
