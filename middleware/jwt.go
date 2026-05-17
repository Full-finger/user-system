package middleware

import (
	"net/http"
	"strings"

	"github.com/full-finger/user-system/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(cfg *config.JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "缺少token",
				})
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "token格式错误",
				})
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				return []byte(cfg.Secret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "token无效或已过期",
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    401,
					"message": "token解析失败",
				})
			}

			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Set("username", claims["username"].(string))
			c.Set("role", claims["role"].(string))

			return next(c)
		}
	}
}

func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("role").(string) != "admin" {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    403,
					"message": "权限不足",
				})
			}
			return next(c)
		}
	}
}
