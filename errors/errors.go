package errors

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"net/http"
)

const (
	UnknownReason = "UNKNOWN_REASON"
	UnknownCode   = 600

	ParamValid     = "PARAM_VALID"
	ParamValidCode = 601

	FormatInvalid     = "FORMAT_INVALID"
	FormatInvalidCode = 602

	Panic     = "PANIC"
	PanicCode = 603

	errStack = "err_stack"

	SupportPackageIsVersion1 = true
)

type Error struct {
	Status
	Surplus interface{} `json:"surplus,omitempty"`
	Err     string      `json:"error,omitempty"`
	clone   bool
	error
}

func (e *Error) Error() string {
	if e.error != nil {
		e.Err = e.error.Error()
	}
	err, _ := json.Marshal(e)
	return string(err)
	//return fmt.Sprintf("code:%d, reason:%s, msg:%v, metadata:%v, surplus:%v, err:%v", e.Code, e.Reason, e.Message, e.Metadata, e.Surplus, e.error)
}

func NewError(code int, reason, msg string) *Error {
	return &Error{
		Status: Status{
			Code:    int32(code),
			Reason:  reason,
			Message: msg,
		},
	}
}

func GetErr(err error) *Error {
	e := &Error{
		Status: Status{
			Code:    UnknownCode,
			Reason:  UnknownReason,
			Message: UnknownReason,
		},
		error: err,
	}
	errors.As(err, &e)
	return e
}

// grpc error

func (e *Error) Unwrap() error {
	return e.error
}

func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

func (e *Error) WithError(cause error) *Error {
	err := clone(e)
	err.error = fmt.Errorf("%+v", cause)
	return err
}

func (e *Error) WithMetadata(md map[string]string) *Error {
	err := clone(e)
	err.Metadata = md
	return err
}

func (e *Error) WithSurplus(surplus interface{}) *Error {
	err := clone(e)
	err.Surplus = surplus
	return err
}

func (e *Error) WithMessage(msg string) *Error {
	err := clone(e)
	err.Message = msg
	return err
}

// write error code to grpc status

func (e *Error) GRPCStatus() *status.Status {
	eInfo := &errdetails.ErrorInfo{
		Reason:   e.Reason,
		Metadata: e.Metadata, // transmit grpc error stack and others by metadata
	}
	if e.error != nil {
		if eInfo.Metadata == nil {
			eInfo.Metadata = map[string]string{}
		}
		eInfo.Metadata[errStack] = fmt.Sprintf("%+v", e.error)
	}
	s, _ := status.New(convertToGRPCCode(int(e.Code)), e.Message).WithDetails(eInfo)
	return s
}

func Code(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return int(FromError(err).Code)
}

func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

func clone(err *Error) *Error {
	if err.clone {
		return err
	}
	err.clone = true
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &Error{
		error: err.error,
		Status: Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
		Surplus: err.Surplus,
	}
}

// convert error to Error

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if !ok {
		return NewError(UnknownCode, UnknownReason, err.Error())
	}
	ret := NewError(
		convertFromGRPCCode(gs.Code()),
		UnknownReason,
		gs.Message(),
	)
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Reason = d.Reason
			ret = ret.WithMetadata(d.Metadata)
			ret.Err = ret.Metadata[errStack]
			delete(ret.Metadata, errStack)
			return ret
		}
	}
	return ret
}
