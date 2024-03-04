package logger

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/trace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type RawJSONFormatter struct {
	*logrus.JSONFormatter
}

func (f *RawJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	if f.DataKey != "" {
		newData := make(logrus.Fields, 4)
		newData[f.DataKey] = data
		data = newData
	}

	fm := make(fieldMap, len(f.FieldMap))
	for fk, fv := range f.FieldMap {
		fm[string(fk)] = fv
	}

	prefixFieldClashes(data, fm, entry.HasCaller())

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	if !f.DisableTimestamp {
		data[fm.resolve(logrus.FieldKeyTime)] = entry.Time.Format(timestampFormat)
	}
	data[fm.resolve(logrus.FieldKeyMsg)] = entry.Message
	data[fm.resolve(logrus.FieldKeyLevel)] = entry.Level.String()
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		}
		if funcVal != "" {
			data[fm.resolve(logrus.FieldKeyFunc)] = funcVal
		}
		if fileVal != "" {
			data[fm.resolve(logrus.FieldKeyFile)] = fileVal
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	value := make(map[string]interface{}, len(data))
	for dk, dv := range data {
		value[dk] = dv
	}

	convert(b, value)
	return b.Bytes(), nil
}

func convert(buf *bytes.Buffer, data map[string]interface{}) {
	buf.WriteByte('{')
	i := 0
	for k, v := range data {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(`"%s":`, k))
		switch val := v.(type) {
		case string:
			buf.WriteString(fmt.Sprintf(`"%s"`, v))
		case map[string]interface{}:
			convert(buf, val)
		default:
			buf.WriteString(fmt.Sprintf("%v", v))
		}
		i++
	}
	buf.WriteByte('}')
	buf.WriteByte('\n')
}

func prefixFieldClashes(data logrus.Fields, fieldMap fieldMap, reportCaller bool) {
	timeKey := fieldMap.resolve(logrus.FieldKeyTime)
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}

	msgKey := fieldMap.resolve(logrus.FieldKeyMsg)
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}

	levelKey := fieldMap.resolve(logrus.FieldKeyLevel)
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}

	logrusErrKey := fieldMap.resolve(logrus.FieldKeyLogrusError)
	if l, ok := data[logrusErrKey]; ok {
		data["fields."+logrusErrKey] = l
		delete(data, logrusErrKey)
	}

	// If reportCaller is not set, 'func' will not conflict.
	if reportCaller {
		funcKey := fieldMap.resolve(logrus.FieldKeyFunc)
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := fieldMap.resolve(logrus.FieldKeyFile)
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
}

type fieldMap map[string]string

func (f fieldMap) resolve(key string) string {
	if k, ok := f[key]; ok {
		return k
	}
	return key
}

type log struct {
	*logrus.Logger
}

func NewLog(opts ...FuncOpts) Logger {
	le := &logEntity{
		name:   "default",
		level:  logrus.DebugLevel,
		levels: logrus.AllLevels,
		formatter: &RawJSONFormatter{
			JSONFormatter: &logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05.000",
			},
		},
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(le)
	}
	l := logrus.StandardLogger()
	l.SetFormatter(le.formatter)
	l.SetLevel(le.level)
	l.SetOutput(le.writer)
	l.SetReportCaller(le.reportCaller)
	l.AddHook(le)
	return &log{Logger: l}
}

func (l *log) Log(ctx context.Context, level uint, fields map[string]interface{}, v ...interface{}) {
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
	l.WithContext(ctx).WithFields(fields).Log(logrusLevel, v)
}

// logrus opt

type logEntity struct {
	name         string
	level        logrus.Level
	levels       []logrus.Level
	formatter    logrus.Formatter
	writer       io.Writer
	writers      map[logrus.Level]io.Writer
	reportCaller bool
}

type FuncOpts func(*logEntity)

func WithSrvName(name string) FuncOpts {
	return func(l *logEntity) {
		l.name = name
	}
}

func WithLevel(level string) FuncOpts {
	return func(l *logEntity) {
		lv, err := logrus.ParseLevel(level)
		if err != nil {
			panic(fmt.Errorf("logrus parse level fail, level:%s, err:%+v", level, err))
		}
		l.level = lv
	}
}

func WithLevels(levels []string) FuncOpts {
	return func(l *logEntity) {
		lvs := make([]logrus.Level, 0, len(levels))
		for _, level := range levels {
			lv, err := logrus.ParseLevel(level)
			if err != nil {
				panic(fmt.Errorf("logrus parse level fail, levle:%s, err:%+v", level, err))
			}
			lvs = append(lvs, lv)
		}
		l.levels = lvs
	}
}

func WithFormatter(formatter logrus.Formatter) FuncOpts {
	return func(l *logEntity) {
		l.formatter = formatter
	}
}

func WithWriter(writer io.Writer) FuncOpts {
	return func(l *logEntity) {
		l.writer = writer
	}
}

func WithDispatcher(dispatcher map[string]io.Writer) FuncOpts {
	return func(l *logEntity) {
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
	return func(l *logEntity) {
		l.reportCaller = caller
	}
}

func (l *logEntity) Levels() []logrus.Level {
	return l.levels
}

func (l *logEntity) Fire(entry *logrus.Entry) error {
	traceID := trace.ExtractTraceID(entry.Context)
	entry.Data[utils.TraceID] = traceID
	entry.Data[utils.LogName] = l.name

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
