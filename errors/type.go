package errors

// BadRequest param / format error
func BadRequest(msg, reason string) *Error {
	return New(400, msg, reason)
}

func IsBadRequest(err error) bool {
	return Code(err) == 400
}

// Unauthorized token invalid / expired
func Unauthorized(msg, reason string) *Error {
	return New(401, msg, reason)
}

func IsUnauthorized(err error) bool {
	return Code(err) == 401
}

// Forbidden token has no rights to access
func Forbidden(msg, reason string) *Error {
	return New(403, msg, reason)
}

func IsForbidden(err error) bool {
	return Code(err) == 403
}

func NotFound(msg, reason string) *Error {
	return New(404, msg, reason)
}

func IsNotFound(err error) bool {
	return Code(err) == 404
}

// InternalServer network / database error
func InternalServer(msg, reason string) *Error {
	return New(500, msg, reason)
}

func IsInternalServer(err error) bool {
	return Code(err) == 500
}

// ServerUnavailable panic
func ServerUnavailable(msg, reason string) *Error {
	return New(503, msg, reason)
}

func IsServerUnavailable(err error) bool {
	return Code(err) == 503
}
