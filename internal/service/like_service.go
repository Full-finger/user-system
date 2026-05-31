package service

import (
	"context"

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
func (s *LikeService) FindLikedPostIDs(ctx context.Context, userID uint, postIDs []uint) (map[uint]bool, error) {
	m, err := s.likeRepo.FindLikedPostIDs(ctx, userID, postIDs)
	if err != nil {
		s.log.Error("批量查询点赞状态失败", zap.Error(err))
		return nil, err
	}
	return m, nil
}
