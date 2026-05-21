package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/router"
	"github.com/full-finger/user-system/internal/service"
	"github.com/full-finger/user-system/pkg/email"
	"github.com/full-finger/user-system/pkg/validator"
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("连接 Redis 失败:", err)
	}

	mailer := email.NewSender(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.From, cfg.SMTP.TLS,
	)

	userSvc := service.NewUserService(db, &cfg.JWT)
	captchaSvc := service.NewCaptchaService(rdb, &cfg.Captcha, mailer)
	userCtrl := controller.NewUserController(userSvc, captchaSvc)

	e := echo.New()
	e.Validator = validator.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Code, map[string]any{"code": appErr.Code, "message": appErr.Message, "data": nil})
			return
		}
		c.JSON(500, map[string]any{"code": 500, "message": "内部错误", "data": nil})
	}
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	router.Setup(e, userCtrl, cfg)

	// Graceful shutdown
	go func() {
		if err := e.Start(cfg.Server.Port); err != nil {
			e.Logger.Info("shutting down:", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
