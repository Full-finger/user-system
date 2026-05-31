package auth

import (
	"github.com/full-finger/user-system/internal/apperror"
	"github.com/labstack/echo/v4"
)

const userContextKey = "user_context"

// UserContext 统一用户上下文，贯穿中间件 → Controller → Service。
type UserContext struct {
	Role     Role
	UserID   uint
	DeviceID string
	Username string
}

func (uc *UserContext) IsGuest() bool         { return uc.Role == RoleGuest }
func (uc *UserContext) IsAuthenticated() bool { return uc.Role > RoleGuest }

// RequireRole 当角色级别不足时返回 Forbidden，否则返回 nil。
func (uc *UserContext) RequireRole(minRole Role) error {
	if uc.Role.Level() < minRole.Level() {
		return apperror.Forbidden("权限不足")
	}
	return nil
}

// GetUserContext 从 echo.Context 提取 UserContext，永远不为 nil。
func GetUserContext(c echo.Context) *UserContext {
	if uc, ok := c.Get(userContextKey).(*UserContext); ok {
		return uc
	}
	return &UserContext{Role: RoleGuest}
}

// SetUserContext 注入 UserContext 到 echo.Context。
func SetUserContext(c echo.Context, uc *UserContext) {
	c.Set(userContextKey, uc)
}
