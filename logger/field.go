package logger

// 日志字段收敛

const (
	Level     = "lv"
	APPID     = "aid"
	Timestamp = "ts"
	Msg       = "msg"
	Module    = "mod"
	Error     = "err"
)

type Field struct {
	Key   string
	Value interface{}
}

func SetField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}
