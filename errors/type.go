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

func LoginFail(msg string) *Error {
	return New(FailLoginCode, msg, FailLogin)
}

func LogoutFail(msg string) *Error {
	return New(FailLogoutCode, msg, FailLogout)
}

func NotFoundData(msg string) *Error {
	return New(DataNotFoundCode, msg, DataNotFound)
}

func IsNotFoundData(err error) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	return e.Reason == DataNotFound && e.Code == DataNotFoundCode
}

func ExistsData(msg string) *Error {
	return New(DataExistsCode, msg, DataExists)
}

func IsExistsData(err error) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	return e.Reason == DataExists && e.Code == DataExistsCode
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

func CreateData(msg string) *Error {
	return New(DataCreateCode, msg, DataCreate)
}

func UpdateData(msg string) *Error {
	return New(DataUpdateCode, msg, DataUpdate)
}

func DeleteData(msg string) *Error {
	return New(DataDeleteCode, msg, DataDelete)
}

func ListData(msg string) *Error {
	return New(DataListCode, msg, DataList)
}
