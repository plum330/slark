package properties

import (
	"github.com/go-slark/slark/encoding"
	"github.com/gookit/properties"
)

type codec struct{}

const Name = "properties"

func init() {
	encoding.RegisterCodec(codec{})
}

func (c codec) Name() string {
	return Name
}

func (c codec) Marshal(v any) ([]byte, error) {
	return properties.Marshal(v)
}

func (c codec) Unmarshal(data []byte, v any) error {
	return properties.Unmarshal(data, v)
}
