package main

import (
	"log"

	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/router"
	"github.com/full-finger/user-system/internal/service"
	customValidator "github.com/full-finger/user-system/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	db.AutoMigrate(&model.User{})

	userSvc := service.NewUserService(db, &cfg.JWT)
	userCtrl := controller.NewUserController(userSvc)

	e := echo.New()
	e.Validator = customValidator.NewValidator()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	router.Setup(e, userCtrl, cfg)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
