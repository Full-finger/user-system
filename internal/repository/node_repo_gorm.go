package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
	"gorm.io/gorm"
)

type nodeRepoGorm struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepoGorm{db: db}
}

func (r *nodeRepoGorm) Create(ctx context.Context, node *model.Node) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *nodeRepoGorm) FindByID(ctx context.Context, id uint) (*model.Node, error) {
	var node model.Node
	if err := r.db.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepoGorm) FindBySlug(ctx context.Context, slug string) (*model.Node, error) {
	var node model.Node
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *nodeRepoGorm) FindAll(ctx context.Context) ([]model.Node, error) {
	var nodes []model.Node
	if err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *nodeRepoGorm) IncrPostCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Node{}).Where("id = ?", id).
		UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error
}

func (r *nodeRepoGorm) DecrPostCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.Node{}).Where("id = ? AND post_count > 0", id).
		UpdateColumn("post_count", gorm.Expr("post_count - 1")).Error
}
