package repository

import "context"

// NodeModeratorRepository 节点版主数据访问接口。
type NodeModeratorRepository interface {
	IsModerator(ctx context.Context, userID, nodeID uint) (bool, error)
}
