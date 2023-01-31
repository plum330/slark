package logger

import (
	"context"
	"github.com/go-slark/slark/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type Log struct {
	*logrus.Logger
}

func NewLog(opts ...FuncOpts) *Log {
	l := &logger{
		srvName: "Default-Server",
		level:   logrus.DebugLevel,
		levels:  logrus.AllLevels,
		formatter: &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			PrettyPrint:     false,
		},
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(l)
	}
	stdLogger := logrus.StandardLogger()
	stdLogger.SetFormatter(l.formatter)
	stdLogger.SetLevel(l.level)
	stdLogger.SetOutput(l.writer)
	stdLogger.SetReportCaller(l.reportCaller)
	stdLogger.AddHook(l)
	return &Log{Logger: stdLogger}
}

func (l *Log) Log(ctx context.Context, level uint, fields map[string]interface{}, msg ...interface{}) {
	var logrusLevel logrus.Level
	switch level {
	case DebugLevel:
		logrusLevel = logrus.DebugLevel
	case InfoLevel:
		logrusLevel = logrus.InfoLevel
	case WarnLevel:
		logrusLevel = logrus.WarnLevel
	case ErrorLevel:
		logrusLevel = logrus.ErrorLevel
	case FatalLevel:
		logrusLevel = logrus.FatalLevel
	case PanicLevel:
		logrusLevel = logrus.PanicLevel
	case TraceLevel:
		logrusLevel = logrus.TraceLevel
	default:
		logrusLevel = logrus.DebugLevel
	}
	l.WithContext(ctx).WithFields(fields).Log(logrusLevel, msg)
}

// logrus opt

type logger struct {
	srvName      string
	level        logrus.Level
	levels       []logrus.Level
	formatter    logrus.Formatter
	writer       io.Writer
	writers      map[logrus.Level]io.Writer
	reportCaller bool
}

type FuncOpts func(*logger)

func WithSrvName(name string) FuncOpts {
	return func(l *logger) {
		l.srvName = name
	}
}

func WithLevel(level string) FuncOpts {
	return func(l *logger) {
		lv, err := logrus.ParseLevel(level)
		if err != nil {
			panic(errors.Errorf("logrus parse level fail, level:%s, err:%+v", level, err))
		}
		l.level = lv
	}
}

func WithLevels(levels []string) FuncOpts {
	return func(l *logger) {
		lvs := make([]logrus.Level, 0, len(levels))
		for _, level := range levels {
			lv, err := logrus.ParseLevel(level)
			if err != nil {
				panic(errors.Errorf("logrus parse level fail, levle:%s, err:%+v", level, err))
			}
			lvs = append(lvs, lv)
		}
		l.levels = lvs
	}
}

func WithFormatter(formatter logrus.Formatter) FuncOpts {
	return func(l *logger) {
		l.formatter = formatter
	}
}

func WithWriter(writer io.Writer) FuncOpts {
	return func(l *logger) {
		l.writer = writer
	}
}

func WithDispatcher(dispatcher map[string]io.Writer) FuncOpts {
	return func(l *logger) {
		l.levels = make([]logrus.Level, 0, len(dispatcher))
		l.writers = make(map[logrus.Level]io.Writer, len(dispatcher))
		maxLevel := logrus.Level(len(logrus.AllLevels))
		for level, writer := range dispatcher {
			lv, err := logrus.ParseLevel(level)
			if err != nil {
				continue
			}

			if maxLevel <= lv {
				continue
			}
			l.writers[lv] = writer
			l.levels = append(l.levels, lv)
		}
	}
}

func WithReportCaller(caller bool) FuncOpts {
	return func(l *logger) {
		l.reportCaller = caller
	}
}

func (l *logger) Levels() []logrus.Level {
	return l.levels
}

func (l *logger) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}
	entry.Data[pkg.TraceID] = ctx.Value(pkg.TraceID)
	entry.Data[pkg.ServerName] = l.srvName

	// 日志统一分发 es mongo kafka
	writer, ok := l.writers[entry.Level]
	if !ok {
		return nil
	}
	eb, err := entry.Bytes()
	if err != nil {
		return err
	}
	_, err = writer.Write(eb)
	return err
}
