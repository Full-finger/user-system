package model

import "time"

// Mention 帖子中的 @提及记录。
type Mention struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PostID    uint      `json:"post_id" gorm:"uniqueIndex:idx_mention_post_user;not null"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex:idx_mention_post_user;not null"`
	Username  string    `json:"username" gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
