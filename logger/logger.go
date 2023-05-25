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

var logger Logger

func init() {
	logger = NewLog()
}

func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}

func Log(ctx context.Context, level uint, fields map[string]interface{}, v ...interface{}) {
	logger.Log(ctx, level, fields, v)
}
