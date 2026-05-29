// Package param 定义请求/响应 DTO。
package param

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
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
	Password string `json:"password" validate:"omitempty,min=6"`
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
