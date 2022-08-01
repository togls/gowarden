package auth

import "net/http"

type RespError struct {
	Kind int
	code int
	msg  string
}

func NewRespError(msg string, o ...int) *RespError {

	e := &RespError{
		msg:  msg,
		code: http.StatusBadRequest,
	}

	if len(o) > 0 {
		e.code = o[0]
	}

	if len(o) > 1 {
		e.Kind = o[1]
	}

	return e
}

func (e RespError) Error() string {
	return e.msg
}

func (e RespError) Response(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.code)
	s := `{
        "Message": ` + e.msg + `,
        "error": "",
        "error_description": "",
        "ValidationErrors": {"": [ msg ]},
        "ErrorModel": {
            "Message": msg,
            "Object": "error"
        },
        "ExceptionMessage": null,
        "ExceptionStackTrace": null,
        "InnerExceptionMessage": null,
        "Object": "error"
    }`
	w.Write([]byte(s))
}

var ErrScopeNotSupported = NewRespError("Scope not supported")
var ErrInvalidPassword = NewRespError("Username or password is incorrect. Try again")
var ErrUserDisabled = NewRespError("This user has been disabled")
