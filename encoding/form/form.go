package form

import (
	"github.com/go-playground/form/v4"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/errors"
	"google.golang.org/protobuf/proto"
	"net/url"
	"reflect"
)

const Name = "x-www-form-urlencoded"

func init() {
	decoder := form.NewDecoder()
	decoder.SetTagName("json")
	encoding.RegisterCodec(&codec{
		decoder: decoder,
	})
}

type codec struct {
	decoder *form.Decoder
}

func (*codec) Marshal(v interface{}) ([]byte, error) {
	values, ok := v.(url.Values)
	if !ok {
		return nil, errors.BadRequest(errors.ParamError, errors.ParamError)
	}
	return []byte(values.Encode()), nil
}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	if m, ok := v.(proto.Message); ok {
		return DecodeValues(m, values)
	}
	if m, ok := rv.Interface().(proto.Message); ok {
		return DecodeValues(m, values)
	}

	return c.decoder.Decode(v, values)
}

func (*codec) Name() string {
	return Name
}
