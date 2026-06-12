package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// CommentLikeRepository 评论点赞数据访问接口。
type CommentLikeRepository interface {
	Exists(ctx context.Context, userID, commentID uint) (bool, error)
	Create(ctx context.Context, like *model.CommentLike) error
	Delete(ctx context.Context, userID, commentID uint) error
	FindLikedCommentIDs(ctx context.Context, userID uint, commentIDs []uint) (map[uint]bool, error)
}
