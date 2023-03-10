package errors

import (
	"google.golang.org/grpc/codes"
	"net/http"
)

const (
	UnknownReason = "UNKNOWN_REASON"
	UnknownCode   = 600

	InvalidParam   = "PARAM_INVALID"
	ParamValidCode = 601

	InvalidFormat     = "FORMAT_INVALID"
	FormatInvalidCode = 602

	Panic     = "SERVER_SLEEPY"
	PanicCode = 603

	Database     = "DATABASE_ERROR"
	DatabaseCode = 604

	Network     = "NETWORK_ERROR"
	NetworkCode = 605

	Denied     = "UNAUTHORIZED"
	DeniedCode = 606

	DataNotFound     = "DATA_NOT_FOUND"
	DataNotFoundCode = 607

	InternalServer     = "INTERNAL_SERVER"
	InternalServerCode = 608

	RequestBad     = "BAD_REQUEST"
	RequestBadCode = 609

	InvalidToken     = "INVALID_TOKEN"
	InvalidTokenCode = 610

	ExpireToken     = "EXPIRE_TOKEN"
	ExpireTokenCode = 611

	FailLogin     = "LOGIN_FAIL"
	FailLoginCode = 612

	FailLogout     = "LOGOUT_FAIL"
	FailLogoutCode = 613

	AccountNotFound     = "ACCOUNT_NOT_FOUND"
	AccountNotFoundCode = 614

	AccountExists     = "ACCOUNT_EXISTS"
	AccountExistsCode = 615

	AccountPasswordError     = "ACCOUNT_PASSWORD_ERROR"
	AccountPasswordErrorCode = 616

	SupportPackageIsVersion1 = true

	ClientClosed = 499
)

func HTTPToGRPCCode(code int) codes.Code {
	switch code {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.Aborted
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	case ClientClosed:
		return codes.Canceled
	}
	return codes.Unknown
}

func GRPCToHTTPCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return ClientClosed
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
