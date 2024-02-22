package errors

import "net/http"

// BadRequest param / format error
func BadRequest(msg, reason string) *Error {
	return New(http.StatusBadRequest, msg, reason)
}

func IsBadRequest(err error) bool {
	return Code(err) == http.StatusBadRequest
}

// Unauthorized token invalid / expired
func Unauthorized(msg, reason string) *Error {
	return New(http.StatusUnauthorized, msg, reason)
}

func IsUnauthorized(err error) bool {
	return Code(err) == http.StatusUnauthorized
}

// Forbidden token has no rights to access
func Forbidden(msg, reason string) *Error {
	return New(http.StatusForbidden, msg, reason)
}

func IsForbidden(err error) bool {
	return Code(err) == http.StatusForbidden
}

func NotFound(msg, reason string) *Error {
	return New(http.StatusNotFound, msg, reason)
}

func IsNotFound(err error) bool {
	return Code(err) == http.StatusNotFound
}

// InternalServer network / database error
func InternalServer(msg, reason string) *Error {
	return New(http.StatusInternalServerError, msg, reason)
}

func IsInternalServer(err error) bool {
	return Code(err) == http.StatusInternalServerError
}

// ServerUnavailable panic
func ServerUnavailable(msg, reason string) *Error {
	return New(http.StatusServiceUnavailable, msg, reason)
}

func IsServerUnavailable(err error) bool {
	return Code(err) == http.StatusServiceUnavailable
}

func ServerTimeout(msg, reason string) *Error {
	return New(http.StatusGatewayTimeout, msg, reason)
}

func IsServerTimeout(err error) bool {
	return Code(err) == http.StatusGatewayTimeout
}
