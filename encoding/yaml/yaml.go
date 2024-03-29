package yaml

import (
	"github.com/go-slark/slark/encoding"
	"gopkg.in/yaml.v3"
)

type codec struct{}

func init() {
	encoding.RegisterCodec(codec{})
}

const Name = "yaml"

func (c codec) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (c codec) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func (c codec) Name() string {
	return Name
}
