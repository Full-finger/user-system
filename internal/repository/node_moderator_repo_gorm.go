package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type nodeModRepoGorm struct {
	db *gorm.DB
}

// NewNodeModeratorRepository 创建基于 GORM 的 NodeModeratorRepository 实现。
func NewNodeModeratorRepository(db *gorm.DB) NodeModeratorRepository {
	return &nodeModRepoGorm{db: db}
}

func (r *nodeModRepoGorm) IsModerator(ctx context.Context, userID, nodeID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.NodeModerator{}).
		Where("user_id = ? AND node_id = ?", userID, nodeID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
