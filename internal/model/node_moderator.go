package model

import "time"

// NodeModerator 节点版主关联表。
type NodeModerator struct {
	NodeID    uint      `json:"node_id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}
