package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type followRepoGorm struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepoGorm{db: db}
}

func (r *followRepoGorm) Create(ctx context.Context, follow *model.Follow) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(follow).Error
}

func (r *followRepoGorm) Delete(ctx context.Context, followerID, followingID uint) error {
	return r.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.Follow{}).Error
}

func (r *followRepoGorm) Exists(ctx context.Context, followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *followRepoGorm) FindFollowers(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64
	db := r.db.WithContext(ctx).Where("following_id = ?", userID)
	if err := db.Model(&model.Follow{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("Follower").Order("id DESC").Offset(offset).Limit(size).Find(&follows).Error; err != nil {
		return nil, 0, err
	}
	return follows, total, nil
}

func (r *followRepoGorm) FindFollowings(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64
	db := r.db.WithContext(ctx).Where("follower_id = ?", userID)
	if err := db.Model(&model.Follow{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("Following").Order("id DESC").Offset(offset).Limit(size).Find(&follows).Error; err != nil {
		return nil, 0, err
	}
	return follows, total, nil
}

func (r *followRepoGorm) FollowingIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ?", userID).Pluck("following_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *followRepoGorm) FollowerIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("following_id = ?", userID).Pluck("follower_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *followRepoGorm) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("following_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *followRepoGorm) CountFollowings(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *followRepoGorm) FindFollowedUserIDs(ctx context.Context, followerID uint, userIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	var followedIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND following_id IN ?", followerID, userIDs).
		Pluck("following_id", &followedIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range followedIDs {
		result[id] = true
	}
	return result, nil
}
