package errors

import (
	"fmt"
	"testing"
)

func BaseError() error {
	return New(599, "base error", "base error").WithMessage("first error")
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
