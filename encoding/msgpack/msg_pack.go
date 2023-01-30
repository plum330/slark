package msgpack

import (
	"github.com/smallfish-root/common-pkg/xencoding"
	"github.com/vmihailenco/msgpack/v5"
)

const Name = "msgpack"

func init() {
	xencoding.RegisterCodec(&codec{})
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
