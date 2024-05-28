package xml

import (
	"encoding/xml"
	"github.com/go-slark/slark/encoding"
)

type codec struct{}

const Name = "xml"

func (c codec) Name() string {
	return Name
}

func (c codec) Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}

func (c codec) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

func init() {
	encoding.RegisterCodec(codec{})
}
