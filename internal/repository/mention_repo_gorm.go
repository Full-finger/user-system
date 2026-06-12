package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type mentionRepoGorm struct {
	db *gorm.DB
}

func NewMentionRepository(db *gorm.DB) MentionRepository {
	return &mentionRepoGorm{db: db}
}

func (r *mentionRepoGorm) CreateBatch(ctx context.Context, mentions []model.Mention) error {
	if len(mentions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&mentions).Error
}

func (r *mentionRepoGorm) FindByPostID(ctx context.Context, postID uint) ([]model.Mention, error) {
	var mentions []model.Mention
	if err := r.db.WithContext(ctx).Preload("User").Where("post_id = ? AND comment_id IS NULL", postID).Find(&mentions).Error; err != nil {
		return nil, err
	}
	return mentions, nil
}

func (r *mentionRepoGorm) FindByCommentIDs(ctx context.Context, commentIDs []uint) (map[uint][]model.Mention, error) {
	result := make(map[uint][]model.Mention)
	if len(commentIDs) == 0 {
		return result, nil
	}
	var mentions []model.Mention
	if err := r.db.WithContext(ctx).Preload("User").Where("comment_id IN ?", commentIDs).Find(&mentions).Error; err != nil {
		return nil, err
	}
	for _, m := range mentions {
		if m.CommentID != nil {
			result[*m.CommentID] = append(result[*m.CommentID], m)
		}
	}
	return result, nil
}
