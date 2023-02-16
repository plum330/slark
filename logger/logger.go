package logger

import "context"

const (
	PanicLevel uint = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Logger interface {
	Log(ctx context.Context, level uint, fields map[string]interface{}, v ...interface{})
}
