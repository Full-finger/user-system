package middleware

import (
	"fmt"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware 基于 Redis 滑动窗口的 IP+路径 限流，Redis 异常时放行。
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
			ip := c.RealIP()
			path := c.Path()
			key := fmt.Sprintf("ratelimit:%s:%s", ip, path)

			now := time.Now().UnixMilli()
			windowMs := window.Milliseconds()

			ctx := c.Request().Context()

			pipe := rdb.Pipeline()
			pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now-windowMs))
			countCmd := pipe.ZCard(ctx, key)
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
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
