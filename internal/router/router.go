// Package router 注册 HTTP 路由。
package router

import (
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// Setup 注册所有 API 路由，分为公开、鉴权、管理员三组。
func Setup(e *echo.Echo, ctrl *controller.UserController, cfg *config.Config, rdb *redis.Client) {
	api := e.Group("/api")

	// 公开路由（IP 限流）
	public := api.Group("")
	public.Use(middleware.RateLimitMiddleware(rdb, &cfg.RateLimit))

	public.GET("/check-username", ctrl.CheckUsername)
	public.POST("/register", ctrl.Register)
	public.POST("/login", ctrl.Login)
	public.POST("/send-code", ctrl.SendCode)
	public.POST("/code-login", ctrl.CodeLogin)

	// 需要鉴权的路由
	auth := api.Group("")
	auth.Use(middleware.JWTMiddleware(&cfg.JWT))

	auth.GET("/profile", ctrl.GetProfile)
	auth.PUT("/profile", ctrl.UpdateProfile)
	auth.PUT("/profile/email", ctrl.BindEmail)

	// 管理员路由
	admin := auth.Group("")
	admin.Use(middleware.AdminOnly())

	admin.GET("/users", ctrl.ListUsers)
	admin.GET("/users/:id", ctrl.GetUser)
	admin.PUT("/users/:id", ctrl.UpdateUser)
	admin.DELETE("/users/:id", ctrl.DeleteUser)
}
