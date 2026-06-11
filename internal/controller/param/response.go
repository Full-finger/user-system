package param

// TimeFormat 统一时间格式常量（RFC 3339 / ISO 8601）。
const TimeFormat = "2006-01-02T15:04:05Z07:00"

// LoginResponse 登录成功响应。
type LoginResponse struct {
	Token string `json:"token"`
}
