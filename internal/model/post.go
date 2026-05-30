package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 帖子模型，支持软删除。
type Post struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"index;not null"`
	NodeID     uint           `json:"node_id" gorm:"index;not null"`
	Title      string         `json:"title" gorm:"size:200;not null"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	LikeCount  int            `json:"like_count" gorm:"default:0"`
	ReplyCount int            `json:"reply_count" gorm:"default:0"`
	ViewCount  int            `json:"view_count" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	User       User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Node       Node           `json:"node,omitempty" gorm:"foreignKey:NodeID"`
}
