package logger

type Field struct {
	Key   string
	Value interface{}
}

// 日志字段收敛

func AppID(aid string) Field {
	return Field{
		Key:   "appid",
		Value: aid,
	}
}

func Module(m string) Field {
	return Field{
		Key:   "module",
		Value: m,
	}
}

func Msg(msg interface{}) Field {
	return Field{
		Key:   "msg",
		Value: msg,
	}
}

func Level(level string) Field {
	return Field{
		Key:   "level",
		Value: level,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
