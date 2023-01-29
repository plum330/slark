package proto

import (
	"github.com/go-slark/slark/encoding"
	"google.golang.org/protobuf/proto"
)

const Name = "proto"

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (codec) Name() string {
	return Name
}
