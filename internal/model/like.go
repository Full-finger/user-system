package model

// Like 点赞模型，联合唯一索引防止重复点赞。
type Like struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"uniqueIndex:idx_user_post;not null"`
	PostID uint `json:"post_id" gorm:"uniqueIndex:idx_user_post;not null"`
	Post   Post `json:"post,omitempty" gorm:"foreignKey:PostID"`
}
