package errors

func BadRequest(reason, msg string) error {
	return New(400, reason, msg)
}

func InternalServer(reason, message string) error {
	return New(500, reason, message)
}
