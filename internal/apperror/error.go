package apperror

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func BadRequest(msg string) *AppError   { return New(400, msg) }
func Unauthorized(msg string) *AppError { return New(401, msg) }
func Forbidden(msg string) *AppError    { return New(403, msg) }
func NotFound(msg string) *AppError     { return New(404, msg) }
func TooMany(msg string) *AppError      { return New(429, msg) }
func Internal(msg string) *AppError     { return New(500, msg) }
