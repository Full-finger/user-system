package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware 统一身份解析中间件，替代原 JWTMiddleware/OptionalJWTMiddleware/AdminOnly。
// 无 token 降级为 Guest，JWT 解析失败也降级为 Guest。
func AuthMiddleware(cfg *config.JWTConfig, guestCfg *config.GuestJWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uc := parseIdentity(c, cfg, guestCfg)
			auth.SetUserContext(c, uc)
			return next(c)
		}
	}
}

func parseIdentity(c echo.Context, cfg *config.JWTConfig, guestCfg *config.GuestJWTConfig) *auth.UserContext {
	deviceID := c.Request().Header.Get("X-Device-ID")

	uc := &auth.UserContext{DeviceID: deviceID}

	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if parsed := parseToken(tokenStr, cfg, guestCfg); parsed != nil {
			return parsed
		}
	}

	// 无 token 或解析失败 → Guest
	if uc.DeviceID == "" {
		uc.DeviceID = randomDeviceID()
	}
	return uc
}

// parseToken 尝试解析用户 JWT 或 Guest JWT，失败返回 nil。
func parseToken(tokenStr string, cfg *config.JWTConfig, guestCfg *config.GuestJWTConfig) *auth.UserContext {
	// 尝试用户 JWT
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err == nil && token.Valid {
		if uc := extractUserClaims(token.Claims); uc != nil {
			return uc
		}
	}

	// 尝试 Guest JWT
	if guestCfg != nil {
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(guestCfg.Secret), nil
		})
		if err == nil && token.Valid {
			if uc := extractGuestClaims(token.Claims); uc != nil {
				return uc
			}
		}
	}

	return nil
}

func extractUserClaims(claims jwt.Claims) *auth.UserContext {
	mc, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	userID, ok := mc["user_id"].(float64)
	if !ok {
		return nil
	}
	username, _ := mc["username"].(string)
	role := auth.RoleUser
	if r, ok := mc["role"].(float64); ok {
		role = auth.Role(int(r))
	}
	return &auth.UserContext{
		Role:     role,
		UserID:   uint(userID),
		Username: username,
	}
}

func extractGuestClaims(claims jwt.Claims) *auth.UserContext {
	mc, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	deviceID, _ := mc["device_id"].(string)
	return &auth.UserContext{
		Role:     auth.RoleGuest,
		DeviceID: deviceID,
	}
}

func randomDeviceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
