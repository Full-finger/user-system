package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// CommentRepository 评论数据访问接口。
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	FindByID(ctx context.Context, id uint) (*model.Comment, error)
	FindByPostID(ctx context.Context, postID uint, page, size int) ([]model.Comment, int64, error)
	FindReplies(ctx context.Context, parentID uint, page, size int) ([]model.Comment, int64, error)
	FindPage(ctx context.Context, keyword string, page, size int) ([]model.Comment, int64, error)
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	CountByPostID(ctx context.Context, postID uint) (int64, error)
	CountReplies(ctx context.Context, parentIDs []uint) (map[uint]int64, error)
	FindCommentIDsByPostID(ctx context.Context, postID uint) ([]uint, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	// FindPageByNodeIDs 管理后台按节点 ID 列表筛选评论（版主只能看管辖节点的评论）。
	FindPageByNodeIDs(ctx context.Context, keyword string, nodeIDs []uint, page, size int) ([]model.Comment, int64, error)
}
