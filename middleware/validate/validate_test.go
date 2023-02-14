package validate

import (
	"context"
	"github.com/go-slark/slark/errors"
	"testing"
)

type validator struct{}

func (v *validator) ValidateAll() error {
	return errors.BadRequest("bad request", "bad request")
}

func (v *validator) Validate() error {
	return errors.BadRequest("bad request --", "bad request --")
}

func TestValidateErr(t *testing.T) {
	_, err := Validate()(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})(context.TODO(), &validator{})
	t.Log(err)
}

func TestValidate(t *testing.T) {
	_, err := Validate()(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})(context.TODO(), 1)
	t.Log(err)
}
