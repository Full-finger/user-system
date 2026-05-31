package auth

import (
	"time"

	"github.com/full-finger/user-system/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateGuestToken 签发 Guest JWT，仅包含 device_id。
func GenerateGuestToken(deviceID string, cfg *config.GuestJWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"device_id": deviceID,
		"role":      RoleGuest.String(),
		"exp":       jwt.NewNumericDate(time.Now().Add(cfg.Expire)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
