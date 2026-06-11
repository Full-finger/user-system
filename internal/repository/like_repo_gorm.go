package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type likeRepoGorm struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepoGorm{db: db}
}

func (r *likeRepoGorm) Create(ctx context.Context, like *model.Like) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(like).Error
}

func (r *likeRepoGorm) Delete(ctx context.Context, userID, postID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Like{}).Error
}

func (r *likeRepoGorm) Exists(ctx context.Context, userID, postID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *likeRepoGorm) FindByUserID(ctx context.Context, userID uint, page, size int) ([]model.Like, int64, error) {
	var likes []model.Like
	var total int64
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if err := db.Model(&model.Like{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("Post").Preload("Post.User").Order("id DESC").Offset(offset).Limit(size).Find(&likes).Error; err != nil {
		return nil, 0, err
	}
	return likes, total, nil
}

func (r *likeRepoGorm) FindLikedPostIDs(ctx context.Context, userID uint, postIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(postIDs))
	if len(postIDs) == 0 {
		return result, nil
	}
	var likedIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.Like{}).
		Where("user_id = ? AND post_id IN ?", userID, postIDs).
		Pluck("post_id", &likedIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range likedIDs {
		result[id] = true
	}
	return result, nil
}
