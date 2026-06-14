package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// NodeRepository 节点数据访问接口。
type NodeRepository interface {
	Create(ctx context.Context, node *model.Node) error
	FindByID(ctx context.Context, id uint) (*model.Node, error)
	FindByIDs(ctx context.Context, ids []uint) ([]model.Node, error)
	FindBySlug(ctx context.Context, slug string) (*model.Node, error)
	FindAll(ctx context.Context) ([]model.Node, error)
	Update(ctx context.Context, id uint, upd NodeUpdate) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	IncrPostCount(ctx context.Context, id uint) error
	DecrPostCount(ctx context.Context, id uint) error
}

// NodeUpdate 节点更新字段，nil 表示不更新。
type NodeUpdate struct {
	Name      *string
	Slug      *string
	Desc      *string
	Color     *string
	Icon      *string
	SortOrder *int
}
