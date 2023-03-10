package errors

func BadRequest(msg, reason string) *Error {
	return New(RequestBadCode, msg, reason)
}

func ServerError(msg, reason string) *Error {
	return New(InternalServerCode, msg, reason)
}

func Unauthorized(msg, reason string) *Error {
	return New(DeniedCode, msg, reason)
}

func ParamInvalid(msg, reason string) *Error {
	return New(ParamValidCode, msg, reason)
}

func FormatInvalid(msg, reason string) *Error {
	return New(FormatInvalidCode, msg, reason)
}

func DatabaseError(msg, reason string) *Error {
	return New(DatabaseCode, msg, reason)
}

func TokenInvalid(msg, reason string) *Error {
	return New(InvalidTokenCode, msg, reason)
}

func TokenExpire(msg, reason string) *Error {
	return New(ExpireTokenCode, msg, reason)
}

func LoginFail(msg, reason string) *Error {
	return New(FailLoginCode, msg, reason)
}

func LogoutFail(msg, reason string) *Error {
	return New(FailLogoutCode, msg, reason)
}

func NotFoundAccount(msg string) *Error {
	return New(AccountNotFoundCode, msg, AccountNotFound)
}

func IsNotFoundAccount(err error) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	return e.Reason == AccountNotFound && e.Code == AccountNotFoundCode
}

func ExistsAccount(msg string) *Error {
	return New(AccountExistsCode, msg, AccountExists)
}

func IsExistsAccount(err error) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	return e.Reason == AccountExists && e.Code == AccountExistsCode
}

func ErrorAccountPassword(msg string) *Error {
	return New(AccountPasswordErrorCode, msg, AccountPasswordError)
}

func IsAccountPasswordError(err error) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	return e.Reason == AccountPasswordError && e.Code == AccountPasswordErrorCode
}

func CreateData(msg, reason string) *Error {
	return New(DataCreateCode, msg, reason)
}

func UpdateData(msg, reason string) *Error {
	return New(DataUpdateCode, msg, reason)
}

func DeleteData(msg, reason string) *Error {
	return New(DataDeleteCode, msg, reason)
}

func ListData(msg, reason string) *Error {
	return New(DataListCode, msg, reason)
}
