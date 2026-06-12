package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type postRepoGorm struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepoGorm{db: db}
}

func (r *postRepoGorm) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepoGorm) FindByID(ctx context.Context, id uint) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).Preload("User").Preload("Node").First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepoGorm) FindByCode(ctx context.Context, code string) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).Preload("User").Preload("Node").Where("code = ?", code).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepoGorm) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Post{}, id).Error
}

func (r *postRepoGorm) FindByUserID(ctx context.Context, userID uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("Node").Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepoGorm) FindPage(ctx context.Context, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Preload("User").Preload("Node").Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepoGorm) FindByNodeID(ctx context.Context, nodeID uint, page, size int, sort string) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	db := r.db.WithContext(ctx).Where("node_id = ?", nodeID)
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC"
	if sort == "replies" {
		orderClause = "reply_count DESC, created_at DESC"
	}

	offset := (page - 1) * size
	if err := db.Preload("User").Preload("Node").Order(orderClause).Offset(offset).Limit(size).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepoGorm) FindByUserIDs(ctx context.Context, userIDs []uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	db := r.db.WithContext(ctx).Where("user_id IN ?", userIDs)
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("Node").Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepoGorm) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postRepoGorm) IncrLikeCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *postRepoGorm) DecrLikeCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

func (r *postRepoGorm) FindAdminPage(ctx context.Context, keyword string, nodeID uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	db := r.db.WithContext(ctx)
	if keyword != "" {
		db = db.Where("title ILIKE ?", "%"+keyword+"%")
	}
	if nodeID > 0 {
		db = db.Where("node_id = ?", nodeID)
	}
	if err := db.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	if err := db.Preload("User").Preload("Node").Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *postRepoGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postRepoGorm) IncrViewCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
