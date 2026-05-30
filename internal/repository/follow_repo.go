package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// FollowRepository 关注数据访问接口。
type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, followerID, followingID uint) error
	Exists(ctx context.Context, followerID, followingID uint) (bool, error)
	FindFollowers(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error)
	FindFollowings(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error)
	FollowingIDs(ctx context.Context, userID uint) ([]uint, error)
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	CountFollowings(ctx context.Context, userID uint) (int64, error)
}
