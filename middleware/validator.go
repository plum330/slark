package middleware

import (
	"context"
	"github.com/go-slark/slark/errors"
)

type Validator interface {
	ValidateAll() error
	Validate() error
}

func Validate() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			v, ok := req.(Validator)
			if ok {
				err := v.ValidateAll()
				if err != nil {
					return nil, errors.BadRequest("param invalid", err.Error()).WithError(err)
				}
			}
			return handler(ctx, req)
		}
	}
}
