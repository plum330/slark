package validate

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
)

type Validator interface {
	ValidateAll() error
	Validate() error
}

func Validate() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			v, ok := req.(Validator)
			if ok {
				err := v.ValidateAll()
				if err != nil {
					return nil, errors.BadRequest(errors.ParamError, err.Error()).WithError(err)
				}
			}
			return handler(ctx, req)
		}
	}
}
