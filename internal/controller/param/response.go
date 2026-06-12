package param

// TimeFormat 统一时间格式常量（RFC 3339 / ISO 8601）。
const TimeFormat = "2006-01-02T15:04:05Z07:00"

// LoginResponse 登录成功响应。
type LoginResponse struct {
	Token string `json:"token"`
}

// AdminStatsResponse 管理后台统计概览。
type AdminStatsResponse struct {
	UserCount    int64 `json:"user_count"`
	PostCount    int64 `json:"post_count"`
	CommentCount int64 `json:"comment_count"`
	NodeCount    int64 `json:"node_count"`
}
