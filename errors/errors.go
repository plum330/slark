package errors

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"io"
	"runtime"
	"strings"
)

const stackDepth = 64

type stack []uintptr

type Error struct {
	Status
	stack stack
	clone bool
	error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	str := e.Reason
	if e.error != nil {
		if str != "" {
			str += " --> "
		}
		str += e.error.Error()
	}
	return str
}

func callers(skip ...int) stack {
	var (
		pcs [stackDepth]uintptr
		n   = 3
	)
	if len(skip) > 0 {
		n += skip[0]
	}
	return pcs[:runtime.Callers(n, pcs[:])]
}

func New(code int, msg, reason string) *Error {
	return &Error{
		Status: Status{
			Code:    int32(code),
			Reason:  reason,
			Message: msg,
		},
		stack: callers(),
	}
}

func Code(err error) int {
	if err == nil {
		return 200
	}
	return int(FromError(err).Code)
}

func Reason(err error) string {
	if err == nil {
		return ""
	}
	return FromError(err).Reason
}

func Message(err error) string {
	if err == nil {
		return ""
	}
	return FromError(err).Message
}

func Metadata(err error) map[string]string {
	var md map[string]string
	if err != nil {
		md = FromError(err).Metadata
	}
	return md
}

func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}

	return &Error{
		error: err,
		stack: callers(),
		Status: Status{
			Message:  Message(err),
			Reason:   text,
			Code:     int32(Code(err)),
			Metadata: Metadata(err),
		},
	}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.error
}

func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

func (e *Error) WithError(cause error) *Error {
	err := clone(e)
	err.error = cause
	se := new(Error)
	if errors.As(cause, &se) {
		err.Metadata = se.Metadata
	}
	return err
}

func (e *Error) WithMetadata(md map[string]string) *Error {
	err := clone(e)
	err.Metadata = md
	return err
}

func (e *Error) WithMessage(msg string) *Error {
	err := clone(e)
	err.Message = msg
	return err
}

func (e *Error) WithReason(reason string) *Error {
	err := clone(e)
	err.Reason = reason
	return err
}

// write error code to grpc status

func (e *Error) GRPCStatus() *status.Status {
	eInfo := &errdetails.ErrorInfo{
		Reason:   e.Reason,
		Metadata: e.Metadata,
	}
	s, _ := status.New(HTTPToGRPCCode(int(e.Code)), e.Message).WithDetails(eInfo)
	return s
}

func clone(err *Error) *Error {
	if err.clone {
		return err
	}
	err.clone = true
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		error: err.error,
		stack: err.stack,
		Status: Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
	}
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if !ok {
		return New(UnknownCode, err.Error(), UnknownReason)
	}
	ret := New(GRPCToHTTPCode(gs.Code()), gs.Message(), UnknownReason)
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Reason = d.Reason
			ret = ret.WithMetadata(d.Metadata)
			return ret
		}
	}
	return ret
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('-'):
			if e.Reason != "" {
				io.WriteString(s, e.Reason)
			} else {
				io.WriteString(s, e.Error())
			}
		case s.Flag('+'):
			if verb == 's' {
				io.WriteString(s, e.Stack())
			} else {
				io.WriteString(s, e.Error()+"\n"+e.Stack())
			}
		default:
			io.WriteString(s, e.Error())
		}
	}
}

type stackInfo struct {
	index int
	msg   string
	lines *list.List
}

type stackLine struct {
	funcName string
	lineNo   string
}

func (e *Error) Stack() string {
	if e == nil {
		return ""
	}

	var (
		err   = e
		index = 1
		infos []*stackInfo
	)
	for err != nil {
		info := &stackInfo{
			index: index,
			msg:   fmt.Sprintf("%-v", err),
		}
		index++
		infos = append(infos, info)
		processStack(err.stack, info)

		if err.error == nil {
			break
		}

		ee, ok := err.error.(*Error)
		if !ok {
			infos = append(infos, &stackInfo{
				index: index,
				msg:   err.error.Error(),
			})
			index++
			break
		}
		err = ee
	}
	removeExtraStackLines(infos)
	return formattingStackInfos(infos)
}

func removeExtraStackLines(infos []*stackInfo) {
	var (
		ok      bool
		mp      = make(map[string]struct{})
		info    *stackInfo
		line    *stackLine
		removes []*list.Element
	)
	for i := len(infos) - 1; i >= 0; i-- {
		info = infos[i]
		if info.lines == nil {
			continue
		}

		l := info.lines.Len()
		for n, e := 0, info.lines.Front(); n < l; n, e = n+1, e.Next() {
			line = e.Value.(*stackLine)
			if _, ok = mp[line.lineNo]; ok {
				removes = append(removes, e)
			} else {
				mp[line.lineNo] = struct{}{}
			}
		}
		if len(removes) > 0 {
			for _, e := range removes {
				info.lines.Remove(e)
			}
		}
		removes = removes[:0]
	}
}

func formattingStackInfos(infos []*stackInfo) string {
	buffer := bytes.NewBuffer(nil)
	for index, info := range infos {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", index+1, info.msg))
		if info.lines != nil && info.lines.Len() > 0 {
			formattingStackLines(buffer, info.lines)
		}
	}
	return buffer.String()
}

func formattingStackLines(buffer *bytes.Buffer, lines *list.List) string {
	var (
		line  *stackLine
		space = "  "
		l     = lines.Len()
	)
	for i, e := 0, lines.Front(); i < l; i, e = i+1, e.Next() {
		line = e.Value.(*stackLine)
		if i >= 9 {
			space = " "
		}
		buffer.WriteString(fmt.Sprintf(
			"   %d).%s%s\n        %s\n",
			i+1, space, line.funcName, line.lineNo,
		))
	}
	return buffer.String()
}

func processStack(pcs stack, info *stackInfo) {
	if pcs == nil {
		return
	}

	for _, pc := range pcs {
		fn := runtime.FuncForPC(pc - 1)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc - 1)
		// TODO TestxxError: /slark --> " "
		// || strings.Contains(file, "go/pkg/mod")
		if strings.Contains(file, "<") || strings.Contains(file, "/slark") {
			continue
		}

		// ignore root paths.
		pathLen := len(rootPath)
		if pathLen != 0 && len(file) >= pathLen && file[0:pathLen] == rootPath {
			continue
		}

		if info.lines == nil {
			info.lines = list.New()
		}

		info.lines.PushBack(&stackLine{
			funcName: fn.Name(),
			lineNo:   fmt.Sprintf(`%s:%d`, file, line),
		})
	}
}

type Stack interface {
	Error() string
	Stack() string
}

func HasStack(err error) bool {
	_, ok := err.(Stack)
	return ok
}

var rootPath string

func init() {
	rootPath = strings.ReplaceAll(runtime.GOROOT(), "\\", "/")
}
