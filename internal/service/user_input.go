package service

// RegisterInput 注册输入。
type RegisterInput struct {
	Username string
	Password string
	Email    string
}

// LoginInput 登录输入，Username 字段支持用户名或邮箱。
type LoginInput struct {
	Username string
	Password string
}

// UpdateInput 用户更新输入，空值字段表示不更新。仅管理员可修改 Role。
type UpdateInput struct {
	Password string
	Role     string
}

// ProfileUpdateInput 普通用户修改个人信息输入，不含 Role 字段。
type ProfileUpdateInput struct {
	Password string
}
