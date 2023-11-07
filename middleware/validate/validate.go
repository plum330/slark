package validate

import (
	"context"
	"fmt"
	"github.com/bufbuild/protovalidate-go"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"google.golang.org/protobuf/proto"
	"strings"
)

type Validator interface {
	ValidateAll() error
	Validate() error
}

var (
	v *protovalidate.Validator
	e error
)

func init() {
	v, e = protovalidate.New()
	if e != nil {
		panic(fmt.Sprintf("init protovalidator error:%+v", e))
	}
}

func Validate() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			validator, ok := req.(Validator)
			if ok {
				err := validator.ValidateAll()
				if err != nil {
					es := err.Error()
					return nil, errors.BadRequest(es, es).WithError(err)
				}
			} else {
				m, o := req.(proto.Message)
				if o {
					err := v.Validate(m)
					if err != nil {
						es := err.Error()
						str := strings.Split(es, " ")
						if len(str) == 6 {
							return nil, errors.BadRequest(str[4], es).WithError(err)
						}
						return nil, errors.BadRequest(es, es).WithError(err)
					}
				}
			}
			return handler(ctx, req)
		}
	}
}
