package main

import (
	"context"
	"log"

	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/router"
	"github.com/full-finger/user-system/internal/service"
	customPkg "github.com/full-finger/user-system/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
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

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("连接 Redis 失败:", err)
	}

	// 初始化邮件发送器
	mailer := customPkg.NewEmailSender(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.From, cfg.SMTP.TLS,
	)

	userSvc := service.NewUserService(db, &cfg.JWT)
	captchaSvc := service.NewCaptchaService(rdb, &cfg.Captcha, mailer)
	userCtrl := controller.NewUserController(userSvc, captchaSvc)

	e := echo.New()
	e.Validator = customPkg.NewValidator()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	router.Setup(e, userCtrl, cfg)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}
