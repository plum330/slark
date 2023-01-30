package msgpack

import (
	"github.com/go-slark/slark/encoding"
	"github.com/vmihailenco/msgpack/v5"
)

const Name = "msgpack"

func init() {
	encoding.RegisterCodec(&codec{})
}

type codec struct{}

func (*codec) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (*codec) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

func (*codec) Name() string {
	return Name
}
