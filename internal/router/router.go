package router

import (
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/middleware"
	"github.com/labstack/echo/v4"
)

func Setup(e *echo.Echo, ctrl *controller.UserController, cfg *config.Config) {
	api := e.Group("/api")

	// 公开路由
	api.POST("/register", ctrl.Register)
	api.POST("/login", ctrl.Login)

	// 需要鉴权的路由
	auth := api.Group("")
	auth.Use(middleware.JWTMiddleware(&cfg.JWT))

	auth.GET("/profile", ctrl.GetProfile)
	auth.PUT("/profile", ctrl.UpdateProfile)

	// 管理员路由
	admin := auth.Group("")
	admin.Use(middleware.AdminOnly())

	admin.GET("/users", ctrl.ListUsers)
	admin.GET("/users/:id", ctrl.GetUser)
	admin.PUT("/users/:id", ctrl.UpdateUser)
	admin.DELETE("/users/:id", ctrl.DeleteUser)
}
