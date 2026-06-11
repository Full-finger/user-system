package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/full-finger/user-system/internal/config"
	applogger "github.com/full-finger/user-system/pkg/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	log := applogger.New(cfg.Log)
	defer log.Sync()

	db, err := initDB(cfg, log)
	if err != nil {
		log.Error("连接数据库失败", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}

	rdb, err := initRedis(cfg, log)
	if err != nil {
		log.Error("连接 Redis 失败", zap.Error(err))
		log.Sync()
		os.Exit(1)
	}

	e := echo.New()
	initApp(e, cfg, db, rdb, log)

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
