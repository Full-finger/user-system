package service

// RegisterInput 注册输入。
type RegisterInput struct {
	Username string
	Password string
	Nickname string
	Email    string
}

// LoginInput 登录输入，Username 字段支持用户名或邮箱。
type LoginInput struct {
	Username string
	Password string
}

// UpdateInput 用户更新输入，空值字段表示不更新。仅管理员可修改 Role。
type UpdateInput struct {
	Password   string
	Nickname   string
	Role       string
	CoverTheme *string // nil 不更新；"" 清空为无封面；非空须为合法主题 key
	Motto      *string // nil 不更新；"" 清空；非空为座右铭内容
}

// ProfileUpdateInput 普通用户修改个人信息输入，不含 Role 字段。
type ProfileUpdateInput struct {
	Password   string
	Nickname   string
	CoverTheme *string // nil 不更新；"" 清空为无封面；非空须为合法主题 key
	Motto      *string // nil 不更新；"" 清空；非空为座右铭内容
}
