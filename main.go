package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

func main() {
	// 连接数据库
	dsn := "host=localhost user=myuser password=mypassword dbname=myapp port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// 自动迁移
	db.AutoMigrate(&User{})

	// 创建一个新的Echo实例
	e := echo.New()

	// 日志中间件
	e.Use(middleware.RequestLogger())

	// 恢复中间件
	e.Use(middleware.Recover())

	e.GET("/hello", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
