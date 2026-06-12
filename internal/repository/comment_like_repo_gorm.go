package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type commentLikeRepoGorm struct {
	db *gorm.DB
}

func NewCommentLikeRepository(db *gorm.DB) CommentLikeRepository {
	return &commentLikeRepoGorm{db: db}
}

func (r *commentLikeRepoGorm) Exists(ctx context.Context, userID, commentID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.CommentLike{}).
		Where("user_id = ? AND comment_id = ?", userID, commentID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *commentLikeRepoGorm) Create(ctx context.Context, like *model.CommentLike) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(like).Error
}

func (r *commentLikeRepoGorm) Delete(ctx context.Context, userID, commentID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&model.CommentLike{}).Error
}

func (r *commentLikeRepoGorm) FindLikedCommentIDs(ctx context.Context, userID uint, commentIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(commentIDs))
	if len(commentIDs) == 0 {
		return result, nil
	}
	var likedIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.CommentLike{}).
		Where("user_id = ? AND comment_id IN ?", userID, commentIDs).
		Pluck("comment_id", &likedIDs).Error; err != nil {
		return nil, err
	}
	for _, id := range likedIDs {
		result[id] = true
	}
	return result, nil
}
