package auth

// Role 角色类型，基于 int 的 6 级层级体系。
type Role int

const (
	RoleGuest        Role = iota // 0 游客
	RoleUser                     // 1 普通用户
	RoleVerifiedUser             // 2 认证用户（预留）
	RoleModerator                // 3 版主
	RoleAdmin                    // 4 管理员
	RoleSuperAdmin               // 5 超级管理员
)

// Level 返回角色级别，用于层级比较。
func (r Role) Level() int { return int(r) }

// String 返回角色名称，用于日志和 API 序列化。
func (r Role) String() string {
	switch r {
	case RoleGuest:
		return "guest"
	case RoleUser:
		return "user"
	case RoleVerifiedUser:
		return "verified_user"
	case RoleModerator:
		return "moderator"
	case RoleAdmin:
		return "admin"
	case RoleSuperAdmin:
		return "super_admin"
	default:
		return "unknown"
	}
}

// roleNames 用于反向查找。
var roleNames = map[string]Role{
	"guest":         RoleGuest,
	"user":          RoleUser,
	"verified_user": RoleVerifiedUser,
	"moderator":     RoleModerator,
	"admin":         RoleAdmin,
	"super_admin":   RoleSuperAdmin,
}

// ParseRole 将字符串解析为 Role，无效值返回 RoleGuest。
func ParseRole(s string) Role {
	if r, ok := roleNames[s]; ok {
		return r
	}
	return RoleGuest
}
