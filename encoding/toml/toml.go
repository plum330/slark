package toml

import (
	"github.com/BurntSushi/toml"
	"github.com/go-slark/slark/encoding"
)

const Name = "toml"

type codec struct{}

func init() {
	encoding.RegisterCodec(codec{})
}

func (c codec) Name() string {
	return Name
}

func (c codec) Marshal(v any) ([]byte, error) {
	return []byte{}, nil
}

func (c codec) Unmarshal(data []byte, v any) error {
	return toml.Unmarshal(data, v)
}
