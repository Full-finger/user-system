package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 帖子评论模型，支持无限嵌套（楼中楼）和软删除。
type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PostID    uint           `json:"post_id" gorm:"index;not null"`
	UserID    uint           `json:"user_id" gorm:"index;not null"`
	ParentID  *uint          `json:"parent_id" gorm:"index"`
	ReplyToID *uint          `json:"reply_to_id"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	LikeCount int            `json:"like_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ReplyTo   *User          `json:"reply_to,omitempty" gorm:"foreignKey:ReplyToID"`
}

// CommentLike 评论点赞记录。
type CommentLike struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex:idx_comment_like_user;not null"`
	CommentID uint      `json:"comment_id" gorm:"uniqueIndex:idx_comment_like_user;not null"`
	CreatedAt time.Time `json:"created_at"`
}
