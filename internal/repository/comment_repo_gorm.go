package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type commentRepoGorm struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepoGorm{db: db}
}

func (r *commentRepoGorm) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepoGorm) FindByID(ctx context.Context, id uint) (*model.Comment, error) {
	var c model.Comment
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *commentRepoGorm) FindByPostID(ctx context.Context, postID uint, page, size int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	db := r.db.WithContext(ctx).Where("post_id = ? AND parent_id IS NULL", postID)
	if err := db.Model(&model.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("ReplyTo").
		Order("id ASC").Offset(offset).Limit(size).Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (r *commentRepoGorm) FindReplies(ctx context.Context, parentID uint, page, size int) ([]model.Comment, int64, error) {
	var replies []model.Comment
	var total int64
	db := r.db.WithContext(ctx).Where("parent_id = ?", parentID)
	if err := db.Model(&model.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("ReplyTo").
		Order("id ASC").Offset(offset).Limit(size).Find(&replies).Error; err != nil {
		return nil, 0, err
	}
	return replies, total, nil
}

func (r *commentRepoGorm) CountByPostID(ctx context.Context, postID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *commentRepoGorm) CountReplies(ctx context.Context, parentIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64)
	if len(parentIDs) == 0 {
		return result, nil
	}
	type row struct {
		ParentID uint
		Count    int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).
		Select("parent_id, COUNT(*) as count").
		Where("parent_id IN ?", parentIDs).
		Group("parent_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.ParentID] = r.Count
	}
	return result, nil
}

func (r *commentRepoGorm) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *commentRepoGorm) FindPage(ctx context.Context, keyword string, page, size int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	db := r.db.WithContext(ctx)
	if keyword != "" {
		db = db.Where("content ILIKE ?", "%"+keyword+"%")
	}
	if err := db.Model(&model.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("ReplyTo").
		Order("created_at DESC").Offset(offset).Limit(size).Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (r *commentRepoGorm) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, id).Error
}

func (r *commentRepoGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *commentRepoGorm) FindCommentIDsByPostID(ctx context.Context, postID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
