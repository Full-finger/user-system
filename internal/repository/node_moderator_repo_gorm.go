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

func (r *nodeModRepoGorm) CreateBatch(ctx context.Context, mods []NodeModeratorCreate) error {
	if len(mods) == 0 {
		return nil
	}
	records := make([]model.NodeModerator, len(mods))
	for i, m := range mods {
		records[i] = model.NodeModerator{NodeID: m.NodeID, UserID: m.UserID}
	}
	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *nodeModRepoGorm) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.NodeModerator{}).Error
}

func (r *nodeModRepoGorm) FindByUserID(ctx context.Context, userID uint) ([]uint, error) {
	var nodeIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.NodeModerator{}).
		Where("user_id = ?", userID).Pluck("node_id", &nodeIDs).Error; err != nil {
		return nil, err
	}
	return nodeIDs, nil
}

func (r *nodeModRepoGorm) FindUserIDsByNodeID(ctx context.Context, nodeID uint) ([]uint, error) {
	var userIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.NodeModerator{}).
		Where("node_id = ?", nodeID).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}
