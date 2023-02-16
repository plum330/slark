package errors

func BadRequest(reason, msg string) *Error {
	return New(400, reason, msg)
}

func InternalServer(reason, message string) *Error {
	return New(500, reason, message)
}
