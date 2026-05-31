package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware 基于 Redis 滑动窗口限流，key 优先级：user_id > device_id > IP。
func RateLimitMiddleware(rdb *redis.Client, cfg *config.RateLimitConfig) echo.MiddlewareFunc {
	window := cfg.Window
	if window <= 0 {
		window = time.Minute
	}
	maxReq := cfg.MaxRequest
	if maxReq <= 0 {
		maxReq = 10
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			uc := auth.GetUserContext(c)
			key := rateLimitKey(uc, c.Path(), c.RealIP())

			now := time.Now().UnixMilli()
			windowMs := window.Milliseconds()
			ctx := c.Request().Context()
			member := randMember()

			pipe := rdb.Pipeline()
			pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now-windowMs))
			countCmd := pipe.ZCard(ctx, key)
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
			pipe.Expire(ctx, key, window)
			_, err := pipe.Exec(ctx)
			if err != nil {
				return next(c)
			}

			if countCmd.Val() >= int64(maxReq) {
				return apperror.TooMany("请求过于频繁，请稍后再试")
			}
			return next(c)
		}
	}
}

func rateLimitKey(uc *auth.UserContext, path, ip string) string {
	if uc.IsAuthenticated() {
		return fmt.Sprintf("rl:u:%d:%s", uc.UserID, path)
	}
	if uc.DeviceID != "" {
		return fmt.Sprintf("rl:d:%s:%s", uc.DeviceID, path)
	}
	return fmt.Sprintf("rl:ip:%s:%s", ip, path)
}

func randMember() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
