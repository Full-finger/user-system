package apperror

import "net/http"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func NewWrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

func BadRequest(msg string) *AppError   { return New(http.StatusBadRequest, msg) }
func Unauthorized(msg string) *AppError { return New(http.StatusUnauthorized, msg) }
func Forbidden(msg string) *AppError    { return New(http.StatusForbidden, msg) }
func NotFound(msg string) *AppError     { return New(http.StatusNotFound, msg) }
func TooMany(msg string) *AppError      { return New(http.StatusTooManyRequests, msg) }
func Internal(msg string) *AppError     { return New(http.StatusInternalServerError, msg) }
