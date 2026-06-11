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
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).Find(&mentions).Error; err != nil {
		return nil, err
	}
	return mentions, nil
}
