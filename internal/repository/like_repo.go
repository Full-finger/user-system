package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// LikeRepository 点赞数据访问接口。
type LikeRepository interface {
	Create(ctx context.Context, like *model.Like) error
	Delete(ctx context.Context, userID, postID uint) error
	Exists(ctx context.Context, userID, postID uint) (bool, error)
	FindByUserID(ctx context.Context, userID uint, page, size int) ([]model.Like, int64, error)
	FindLikedPostIDs(ctx context.Context, userID uint, postIDs []uint) (map[uint]bool, error)
	CountReceivedLikesByUserID(ctx context.Context, userID uint) (int64, error)
}
