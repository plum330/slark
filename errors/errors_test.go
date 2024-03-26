package errors

import (
	"errors"
	"fmt"
	"testing"
)

func BaseError() error {
	return New(599, "base error", "base error").WithMessage("first error").
		WithError(errors.New("with errors")).WithReason("77777").
		WithMetadata(map[string]string{"meta": "uuuuuuuu"})
}

func WrapError() error {
	return Wrap(BaseError(), "wrap error")
}

func ReWrapError() error {
	return Wrap(WrapError(), "re wrap error")
}

func TestBaseError(t *testing.T) {
	fmt.Printf("%+v\n", BaseError())
}

func TestWrapError(t *testing.T) {
	fmt.Printf("%+v\n", WrapError())
}

func TestReWrapError(t *testing.T) {
	fmt.Printf("%+v\n", ReWrapError())
}
