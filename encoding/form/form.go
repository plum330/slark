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

	m, ok := v.(proto.Message)
	if ok {
		return parse(m, values)
	}

	value := reflect.ValueOf(v)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}

	m, ok = value.Interface().(proto.Message)
	if ok {
		return parse(m, values)
	}

	return c.decoder.Decode(v, values)
}

func (*codec) Name() string {
	return Name
}
