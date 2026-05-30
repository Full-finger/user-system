package model

import "time"

// Node 节点（相当于贴吧的"吧"），每个帖子必须属于一个节点。
type Node struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null;size:50"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null;size:50"`
	Desc      string    `json:"desc" gorm:"size:200"`
	Color     string    `json:"color" gorm:"size:7"`
	Icon      string    `json:"icon" gorm:"size:50"`
	PostCount int       `json:"post_count" gorm:"default:0"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
