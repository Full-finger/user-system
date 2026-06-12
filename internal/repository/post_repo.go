package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// PostRepository 帖子数据访问接口。
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id uint) (*model.Post, error)
	FindByCode(ctx context.Context, code string) (*model.Post, error)
	Delete(ctx context.Context, id uint) error
	FindByUserID(ctx context.Context, userID uint, page, size int) ([]model.Post, int64, error)
	FindPage(ctx context.Context, page, size int) ([]model.Post, int64, error)
	FindAdminPage(ctx context.Context, keyword string, nodeID uint, page, size int) ([]model.Post, int64, error)
	FindByNodeID(ctx context.Context, nodeID uint, page, size int, sort string) ([]model.Post, int64, error)
	FindByUserIDs(ctx context.Context, userIDs []uint, page, size int) ([]model.Post, int64, error)
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	IncrLikeCount(ctx context.Context, id uint) error
	DecrLikeCount(ctx context.Context, id uint) error
	IncrViewCount(ctx context.Context, id uint) error
}
