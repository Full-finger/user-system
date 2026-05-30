package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// NodeRepository 节点数据访问接口。
type NodeRepository interface {
	Create(ctx context.Context, node *model.Node) error
	FindByID(ctx context.Context, id uint) (*model.Node, error)
	FindBySlug(ctx context.Context, slug string) (*model.Node, error)
	FindAll(ctx context.Context) ([]model.Node, error)
	IncrPostCount(ctx context.Context, id uint) error
	DecrPostCount(ctx context.Context, id uint) error
}
