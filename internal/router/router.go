// Package router 注册 HTTP 路由。
package router

import (
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// Setup 注册所有 API 路由，分为公开、可选鉴权、鉴权、管理员四组。
func Setup(e *echo.Echo, userCtrl *controller.UserController, postCtrl *controller.PostController, cfg *config.Config, rdb *redis.Client) {
	api := e.Group("/api")

	// ── 公开路由（IP 限流） ──────────────────────────────
	public := api.Group("")
	public.Use(middleware.RateLimitMiddleware(rdb, &cfg.RateLimit))

	// 用户认证
	public.GET("/check-username", userCtrl.CheckUsername)
	public.POST("/register", userCtrl.Register)
	public.POST("/login", userCtrl.Login)
	public.POST("/send-code", userCtrl.SendCode)
	public.POST("/code-login", userCtrl.CodeLogin)

	// 节点（公开只读，不需要 liked/followed 状态）
	public.GET("/nodes", postCtrl.ListNodes)
	public.GET("/nodes/:id", postCtrl.GetNode)

	// ── 可选鉴权路由（IP 限流 + 可选 JWT） ──────────────
	// 带 token 时返回个性化数据（liked/followed），不带时正常访问
	optAuth := api.Group("")
	optAuth.Use(middleware.RateLimitMiddleware(rdb, &cfg.RateLimit))
	optAuth.Use(middleware.OptionalJWTMiddleware(&cfg.JWT))

	// 节点帖子
	optAuth.GET("/nodes/:id/posts", postCtrl.ListNodePosts)

	// 帖子（带 liked 状态）
	optAuth.GET("/posts", postCtrl.ListPosts)
	optAuth.GET("/posts/:id", postCtrl.GetPost)

	// 用户相关（带 followed/liked 状态），使用用户名作为标识
	optAuth.GET("/users/:username/posts", postCtrl.ListUserPosts)
	optAuth.GET("/users/:username/likes", postCtrl.ListLikedPosts)
	optAuth.GET("/users/:username/followers", postCtrl.GetFollowers)
	optAuth.GET("/users/:username/followings", postCtrl.GetFollowings)
	optAuth.GET("/users/:username", postCtrl.GetUserProfile)

	// ── 需要鉴权的路由 ───────────────────────────────────
	auth := api.Group("")
	auth.Use(middleware.JWTMiddleware(&cfg.JWT))

	auth.GET("/profile", userCtrl.GetProfile)
	auth.PUT("/profile", userCtrl.UpdateProfile)
	auth.PUT("/profile/email", userCtrl.BindEmail)

	// 帖子 — 写操作
	auth.POST("/posts", postCtrl.CreatePost)
	auth.DELETE("/posts/:id", postCtrl.DeletePost)
	auth.PUT("/posts/:id/like", postCtrl.ToggleLike)
	auth.GET("/feed", postCtrl.ListFeed)

	// 关注
	auth.PUT("/users/:username/follow", postCtrl.ToggleFollow)

	// ── 管理员路由 ──────────────────────────────────────
	admin := auth.Group("/admin")
	admin.Use(middleware.AdminOnly())

	admin.GET("/users", userCtrl.ListUsers)
	admin.GET("/users/:id", userCtrl.GetUser)
	admin.PUT("/users/:id", userCtrl.UpdateUser)
	admin.DELETE("/users/:id", userCtrl.DeleteUser)
	admin.DELETE("/posts/:id", postCtrl.DeletePost)
}
