package repository

import (
	"context"

	"github.com/full-finger/user-system/internal/model"
)

// MentionRepository 提及数据访问接口。
type MentionRepository interface {
	CreateBatch(ctx context.Context, mentions []model.Mention) error
	FindByPostID(ctx context.Context, postID uint) ([]model.Mention, error)
	FindByCommentIDs(ctx context.Context, commentIDs []uint) (map[uint][]model.Mention, error)
}
