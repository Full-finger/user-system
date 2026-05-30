// Package model 定义数据模型。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型，支持软删除。
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null;size:30"`
	Nickname  string         `json:"nickname" gorm:"size:50"`
	Password  string         `json:"-" gorm:"not null"`
	Email     *string        `json:"email" gorm:"uniqueIndex;size:100"`
	Role      string         `json:"role" gorm:"default:user;size:20"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
