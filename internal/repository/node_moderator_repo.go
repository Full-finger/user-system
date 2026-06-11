package repository

import "context"

// NodeModeratorRepository 节点版主数据访问接口。
type NodeModeratorRepository interface {
	IsModerator(ctx context.Context, userID, nodeID uint) (bool, error)
	CreateBatch(ctx context.Context, mods []NodeModeratorCreate) error
	DeleteByUserID(ctx context.Context, userID uint) error
	FindByUserID(ctx context.Context, userID uint) ([]uint, error)
}

// NodeModeratorCreate 创建版主绑定的参数。
type NodeModeratorCreate struct {
	NodeID uint
	UserID uint
}
