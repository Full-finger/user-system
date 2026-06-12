package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/config"
	"github.com/full-finger/user-system/internal/controller"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/full-finger/user-system/internal/router"
	"github.com/full-finger/user-system/internal/service"
	"github.com/full-finger/user-system/pkg/email"
	"github.com/full-finger/user-system/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if cfg.Server.Env != "production" {
		db.AutoMigrate(&model.User{}, &model.Post{}, &model.Like{}, &model.Follow{}, &model.Node{}, &model.Mention{}, &model.NodeModerator{}, &model.Comment{}, &model.CommentLike{})
	} else {
		log.Warn("生产环境跳过 AutoMigrate，请使用迁移工具管理数据库结构")
	}
	return db, nil
}

func initRedis(cfg *config.Config, log *zap.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

// initApp 完成依赖注入、Echo 配置和路由注册。
func initApp(e *echo.Echo, cfg *config.Config, db *gorm.DB, rdb *redis.Client, log *zap.Logger) {
	// --- 依赖注入 ---
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
	commentRepo := repository.NewCommentRepository(db)
	commentLikeRepo := repository.NewCommentLikeRepository(db)

	userSvc := service.NewUserService(userRepo, nodeModRepo, &cfg.JWT, log)
	captchaSvc := service.NewCaptchaService(rdb, &cfg.Captcha, mailer, log)
	nodeSvc := service.NewNodeService(nodeRepo, userRepo, mentionRepo, log)
	likeSvc := service.NewLikeService(likeRepo, log)
	postSvc := service.NewPostService(postRepo, likeRepo, likeSvc, nodeRepo, nodeModRepo, nodeSvc, db, log)
	followSvc := service.NewFollowService(followRepo, userRepo, postRepo, log)
	commentSvc := service.NewCommentService(commentRepo, commentLikeRepo, postRepo, mentionRepo, nodeSvc, db, log)

	// 种子数据
	nodeSvc.SeedNodes(context.Background())
	userSvc.SeedAdmin(context.Background(), &cfg.Admin)

	userCtrl := controller.NewUserController(userSvc, captchaSvc, &cfg.GuestJWT, postRepo, commentRepo, likeRepo, nodeRepo)
	postCtrl := controller.NewPostController(postSvc, nodeSvc, followSvc, log)
	nodeCtrl := controller.NewNodeController(nodeSvc, postSvc)
	followCtrl := controller.NewFollowController(followSvc)
	commentCtrl := controller.NewCommentController(commentSvc, log)
	mentionCtrl := controller.NewMentionController(userRepo, followRepo, nodeModRepo)

	// --- Echo 配置 ---
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

	router.Setup(e, userCtrl, postCtrl, nodeCtrl, followCtrl, commentCtrl, mentionCtrl, cfg, rdb)
}
