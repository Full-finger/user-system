// Package param 定义请求/响应 DTO。
package param

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,password"`
	Nickname string `json:"nickname" validate:"omitempty,nickname"`
	Email    string `json:"email" validate:"required,email"`
	Code     string `json:"code" validate:"required,len=6"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateRequest 用户更新请求，所有字段可选。
type UpdateRequest struct {
	Password string `json:"password" validate:"omitempty,password"`
	Nickname string `json:"nickname" validate:"omitempty,nickname"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

// SendCodeRequest 发送验证码请求。
type SendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// CodeLoginRequest 验证码登录请求。
type CodeLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// BindEmailRequest 绑定邮箱请求。
type BindEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

// CreatePostRequest 发帖请求。
type CreatePostRequest struct {
	NodeID  uint   `json:"node_id" validate:"required"`
	Title   string `json:"title" validate:"required,min=1,max=200"`
	Content string `json:"content" validate:"required,min=1"`
}
