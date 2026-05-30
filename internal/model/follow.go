package model

// Follow 关注模型，联合唯一索引防止重复关注。
type Follow struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	FollowerID  uint `json:"follower_id" gorm:"uniqueIndex:idx_follower_following;not null"`
	FollowingID uint `json:"following_id" gorm:"uniqueIndex:idx_follower_following;not null"`
	Follower    User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	Following   User `json:"following,omitempty" gorm:"foreignKey:FollowingID"`
}
