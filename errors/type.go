package errors

func BadRequest(reason, msg string) *Error {
	return NewError(400, reason, msg)
}

func InternalServer(reason, message string) *Error {
	return NewError(500, reason, message)
}
