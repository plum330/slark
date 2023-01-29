package encoding

import (
	"strings"
)

// abstract serialization / deserialization

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

var codecs = make(map[string]Codec)

func RegisterCodec(codec Codec) {
	if codec == nil || len(codec.Name()) == 0 {
		panic("cannot register nil  or empty name Codec")
	}

	codecs[strings.ToLower(codec.Name())] = codec
}

func GetCodec(codecType string) Codec {
	return codecs[codecType]
}
