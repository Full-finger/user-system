package router

import (
	"net/http"

	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// Setup 注册所有 API 路由，统一走 AuthMiddleware + 限流，权限判断下沉到 Service 层。
func Setup(e *echo.Echo, userCtrl *controller.UserController, postCtrl *controller.PostController, nodeCtrl *controller.NodeController, followCtrl *controller.FollowController, commentCtrl *controller.CommentController, cfg *config.Config, rdb *redis.Client) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware(&cfg.JWT, &cfg.GuestJWT))
	api.Use(middleware.RateLimitMiddleware(rdb, &cfg.RateLimit))

	// 公开
	api.POST("/guest-token", userCtrl.GuestToken)
	api.GET("/check-username", userCtrl.CheckUsername)
	api.POST("/register", userCtrl.Register)
	api.POST("/login", userCtrl.Login)
	api.POST("/send-code", userCtrl.SendCode)
	api.POST("/code-login", userCtrl.CodeLogin)
	api.GET("/nodes", nodeCtrl.ListNodes)
	api.GET("/nodes/:id", nodeCtrl.GetNode)

	// 可个性化（Guest 正常访问，登录用户返回 liked/followed）
	api.GET("/nodes/:id/posts", nodeCtrl.ListNodePosts)
	api.GET("/posts", postCtrl.ListPosts)
	api.GET("/posts/:code", postCtrl.GetPost)
	api.GET("/posts/:code/comments", commentCtrl.ListComments)
	api.GET("/comments/:id/replies", commentCtrl.ListReplies)
	api.GET("/users/:username/posts", postCtrl.ListUserPosts)
	api.GET("/users/:username/likes", postCtrl.ListLikedPosts)
	api.GET("/users/:username/followers", followCtrl.GetFollowers)
	api.GET("/users/:username/followings", followCtrl.GetFollowings)
	api.GET("/users/:username", followCtrl.GetUserProfile)

	// 需登录（Service 层 RequireRole(User)）
	api.GET("/profile", userCtrl.GetProfile)
	api.PUT("/profile", userCtrl.UpdateProfile)
	api.PUT("/profile/email", userCtrl.BindEmail)
	api.POST("/posts", postCtrl.CreatePost)
	api.DELETE("/posts/:code", postCtrl.DeletePost)
	api.PUT("/posts/:code/like", postCtrl.ToggleLike)
	api.POST("/posts/:code/comments", commentCtrl.CreateComment)
	api.PUT("/comments/:id/like", commentCtrl.ToggleCommentLike)
	api.GET("/feed", postCtrl.ListFeed)
	api.PUT("/users/:username/follow", followCtrl.ToggleFollow)

	// 管理（Service 层 RequireRole(Admin/SuperAdmin)）
	api.GET("/admin/users", userCtrl.ListUsers)
	api.GET("/admin/users/:id", userCtrl.GetUser)
	api.PUT("/admin/users/:id", userCtrl.UpdateUser)
	api.DELETE("/admin/users/:id", userCtrl.DeleteUser)
	api.POST("/admin/moderators", userCtrl.AppointModerator)
	api.DELETE("/admin/posts/:code", postCtrl.AdminDeletePost)
}
