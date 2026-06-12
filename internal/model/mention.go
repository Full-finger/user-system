package model

import "time"

// Mention 帖子/评论中的 @提及记录。
type Mention struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PostID    uint      `json:"post_id" gorm:"index;not null"`
	CommentID *uint     `json:"comment_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Username  string    `json:"username" gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
