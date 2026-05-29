// Package middleware 提供 Echo 中间件：JWT 鉴权、权限控制、限流。
package middleware

import (
	"fmt"
	"strings"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware 解析 Bearer token，将 user_id/username/role 注入 Context。
func JWTMiddleware(cfg *config.JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return apperror.Unauthorized("缺少token")
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return apperror.Unauthorized("token格式错误")
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(cfg.Secret), nil
			})
			if err != nil || !token.Valid {
				return apperror.Unauthorized("token无效或已过期")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return apperror.Unauthorized("token解析失败")
			}

			userID, ok := claimFloat(claims, "user_id")
			if !ok {
				return apperror.Unauthorized("token解析失败")
			}
			username, ok := claimString(claims, "username")
			if !ok {
				return apperror.Unauthorized("token解析失败")
			}
			role, ok := claimString(claims, "role")
			if !ok {
				return apperror.Unauthorized("token解析失败")
			}

			c.Set("user_id", uint(userID))
			c.Set("username", username)
			c.Set("role", role)

			return next(c)
		}
	}
}

// AdminOnly 限制仅 admin 角色可访问，需在 JWTMiddleware 之后使用。
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if role != "admin" {
				return apperror.Forbidden("权限不足")
			}
			return next(c)
		}
	}
}

func claimFloat(claims jwt.MapClaims, key string) (float64, bool) {
	v, ok := claims[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

func claimString(claims jwt.MapClaims, key string) (string, bool) {
	v, ok := claims[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
