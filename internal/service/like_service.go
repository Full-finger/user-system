package service

import (
	"context"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
)

// LikeService 点赞查询业务服务。
type LikeService struct {
	likeRepo repository.LikeRepository
	log      *zap.Logger
}

func NewLikeService(likeRepo repository.LikeRepository, log *zap.Logger) *LikeService {
	return &LikeService{likeRepo: likeRepo, log: log}
}

// FindLikedPostIDs 批量查询当前用户对哪些帖子点了赞，返回 postID → bool 映射。
func (s *LikeService) FindLikedPostIDs(ctx context.Context, uc *auth.UserContext, postIDs []uint) (map[uint]bool, error) {
	if uc.IsGuest() || len(postIDs) == 0 {
		return map[uint]bool{}, nil
	}
	m, err := s.likeRepo.FindLikedPostIDs(ctx, uc.UserID, postIDs)
	if err != nil {
		s.log.Error("批量查询点赞状态失败", zap.Error(err))
		return nil, apperror.InternalWrap("查询点赞状态失败", err)
	}
	return m, nil
}
