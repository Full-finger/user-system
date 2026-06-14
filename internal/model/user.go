// Package model 定义数据模型。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型，支持软删除。
type User struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	Username string  `json:"username" gorm:"uniqueIndex;not null;size:30"`
	Nickname string  `json:"nickname" gorm:"size:50"`
	Password string  `json:"-" gorm:"not null"`
	Email    *string `json:"email" gorm:"uniqueIndex;size:100"`
	Role     int     `json:"-" gorm:"default:1;not null"`
	// CoverTheme 个人主页封面主题，空（NULL 或空串）表示无封面。
	CoverTheme *string `json:"cover_theme,omitempty" gorm:"size:20"`
	// Motto 座右铭，展示在个人主页头像下方，空（NULL）表示未设置。
	Motto     *string        `json:"motto,omitempty" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
