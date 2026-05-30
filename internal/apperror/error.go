// Package apperror 提供统一的业务错误类型，所有业务层错误应使用本包构造。
package apperror

import "net/http"

// AppError 携带 HTTP 状态码的业务错误。
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

// NewWrap 包装一个原始错误，保留错误链。
func NewWrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

func BadRequest(msg string) *AppError   { return New(http.StatusBadRequest, msg) }
func Unauthorized(msg string) *AppError { return New(http.StatusUnauthorized, msg) }
func Forbidden(msg string) *AppError    { return New(http.StatusForbidden, msg) }
func NotFound(msg string) *AppError     { return New(http.StatusNotFound, msg) }
func TooMany(msg string) *AppError      { return New(http.StatusTooManyRequests, msg) }
func Internal(msg string) *AppError     { return New(http.StatusInternalServerError, msg) }
