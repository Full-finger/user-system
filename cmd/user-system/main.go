package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/full-finger/user-system/internal/router"
	"github.com/full-finger/user-system/internal/service"
	"github.com/full-finger/user-system/pkg/email"
	applogger "github.com/full-finger/user-system/pkg/logger"
	"github.com/full-finger/user-system/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 初始化配置 → 日志 → 数据库 → Redis → 依赖注入 → 路由 → 启动
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	log := applogger.New(cfg.Log)
	defer log.Sync()

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Error("连接数据库失败", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}
	if cfg.Server.Env != "production" {
		db.AutoMigrate(&model.User{}, &model.Post{}, &model.Like{}, &model.Follow{}, &model.Node{}, &model.Mention{}, &model.NodeModerator{})
	} else {
		log.Warn("生产环境跳过 AutoMigrate，请使用迁移工具管理数据库结构")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Error("连接 Redis 失败", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}

	mailer := email.NewSender(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.From, cfg.SMTP.TLS, cfg.SMTP.Auth,
	)

	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	followRepo := repository.NewFollowRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	nodeModRepo := repository.NewNodeModeratorRepository(db)
	mentionRepo := repository.NewMentionRepository(db)

	userSvc := service.NewUserService(userRepo, nodeModRepo, &cfg.JWT, log)
	captchaSvc := service.NewCaptchaService(rdb, &cfg.Captcha, mailer, log)
	nodeSvc := service.NewNodeService(nodeRepo, userRepo, mentionRepo, log)
	nodeSvc.SeedNodes(context.Background())
	userSvc.SeedAdmin(context.Background(), &cfg.Admin)
	likeSvc := service.NewLikeService(likeRepo, log)
	postSvc := service.NewPostService(postRepo, likeRepo, likeSvc, nodeRepo, nodeModRepo, nodeSvc, db, log)
	followSvc := service.NewFollowService(followRepo, userRepo, postRepo, log)

	userCtrl := controller.NewUserController(userSvc, captchaSvc, &cfg.GuestJWT)
	postCtrl := controller.NewPostController(postSvc, nodeSvc, followSvc)
	nodeCtrl := controller.NewNodeController(nodeSvc, postSvc)
	followCtrl := controller.NewFollowController(followSvc)

	e := echo.New()
	e.Validator = validator.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Code, map[string]any{"code": appErr.Code, "message": appErr.Message, "data": nil})
			return
		}
		log.Error("未处理错误", zap.Error(err), zap.String("path", c.Request().URL.Path))
		c.JSON(500, map[string]any{"code": 500, "message": "内部错误", "data": nil})
	}
	corsOrigins := cfg.Server.CORSOrigins
	if len(corsOrigins) == 0 {
		if cfg.Server.Env == "production" {
			log.Warn("生产环境未配置 cors_origins，CORS 将拒绝所有跨域请求")
		}
		corsOrigins = []string{"http://localhost:5173"}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Device-ID"},
	}))
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	router.Setup(e, userCtrl, postCtrl, nodeCtrl, followCtrl, cfg, rdb)

	log.Info("服务启动", zap.String("port", cfg.Server.Port))
	go func() {
		if err := e.Start(cfg.Server.Port); err != nil {
			log.Info("服务关闭", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Error("服务关闭失败", zap.Error(err))
	}
	log.Info("服务已停止")
}
