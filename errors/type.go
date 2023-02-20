package errors

func BadRequest(msg, reason string) *Error {
	return New(400, msg, reason)
}

func InternalServer(msg, reason string) *Error {
	return New(500, msg, reason)
}

func Unauthorized(msg, reason string) *Error {
	return New(401, msg, reason)
}

func ParamInvalid(msg, reason string) *Error {
	return New(ParamValidCode, msg, reason)
}

func FormatInvalid(msg, reason string) *Error {
	return New(FormatInvalidCode, msg, reason)
}
