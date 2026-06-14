// Package service 实现统计聚合业务逻辑。
package service

import (
	"context"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
)

// StatsService 统计聚合服务，负责跨领域的统计查询。
type StatsService struct {
	userRepo    repository.UserRepository
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	likeRepo    repository.LikeRepository
	nodeRepo    repository.NodeRepository
	nodeModRepo repository.NodeModeratorRepository
	followRepo  repository.FollowRepository
	log         *zap.Logger
}

// NewStatsService 创建统计聚合服务实例。
func NewStatsService(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	likeRepo repository.LikeRepository,
	nodeRepo repository.NodeRepository,
	nodeModRepo repository.NodeModeratorRepository,
	followRepo repository.FollowRepository,
	log *zap.Logger,
) *StatsService {
	return &StatsService{
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
		nodeRepo:    nodeRepo,
		nodeModRepo: nodeModRepo,
		followRepo:  followRepo,
		log:         log,
	}
}

// AdminStatsResult 管理后台统计结果。
type AdminStatsResult struct {
	UserCount    int64
	PostCount    int64
	CommentCount int64
	NodeCount    int64
}

// AdminStats 管理后台统计概览，需要 Admin 及以上权限。
func (s *StatsService) AdminStats(ctx context.Context, uc *auth.UserContext) (*AdminStatsResult, error) {
	if err := uc.RequireRole(auth.RoleAdmin); err != nil {
		return nil, err
	}
	userCount, err := s.userRepo.Count(ctx)
	if err != nil {
		s.log.Error("查询用户总数失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	postCount, err := s.postRepo.Count(ctx)
	if err != nil {
		s.log.Error("查询帖子总数失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	commentCount, err := s.commentRepo.Count(ctx)
	if err != nil {
		s.log.Error("查询评论总数失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	nodeCount, err := s.nodeRepo.Count(ctx)
	if err != nil {
		s.log.Error("查询节点总数失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return &AdminStatsResult{
		UserCount:    userCount,
		PostCount:    postCount,
		CommentCount: commentCount,
		NodeCount:    nodeCount,
	}, nil
}

// ProfileStatsResult 用户资料统计结果。
type ProfileStatsResult struct {
	PostCount      int64
	CommentCount   int64
	LikeCount      int64
	FollowerCount  int64
	FollowingCount int64
	LikedCount     int64
}

// GetProfileStats 获取指定用户的资料统计信息（帖子数、评论数、获赞数、粉丝数、关注数、点赞数）。
func (s *StatsService) GetProfileStats(ctx context.Context, userID uint) (*ProfileStatsResult, error) {
	postCount, err := s.postRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询用户帖子数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	commentCount, err := s.commentRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询用户评论数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	likeCount, err := s.likeRepo.CountReceivedLikesByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询用户获赞数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	followerCount, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		s.log.Error("查询粉丝数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	followingCount, err := s.followRepo.CountFollowings(ctx, userID)
	if err != nil {
		s.log.Error("查询关注数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	likedCount, err := s.likeRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询用户点赞数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return &ProfileStatsResult{
		PostCount:      postCount,
		CommentCount:   commentCount,
		LikeCount:      likeCount,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		LikedCount:     likedCount,
	}, nil
}

// UserStatsResult 用户公开统计结果。
type UserStatsResult struct {
	PostCount      int64
	FollowerCount  int64
	FollowingCount int64
	LikeCount      int64
}

// GetUserStats 获取用户的公开统计信息（帖子数、粉丝数、关注数）。
func (s *StatsService) GetUserStats(ctx context.Context, userID uint) (*UserStatsResult, error) {
	postCount, err := s.postRepo.CountByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询用户帖子数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	followerCount, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		s.log.Error("查询粉丝数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	followingCount, err := s.followRepo.CountFollowings(ctx, userID)
	if err != nil {
		s.log.Error("查询关注数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	likeCount, err := s.likeRepo.CountReceivedLikesByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询获赞数失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return &UserStatsResult{
		PostCount:      postCount,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		LikeCount:      likeCount,
	}, nil
}

// GetMyModeratedNodes 获取当前用户版主管辖的节点列表（需版主及以上权限）。
func (s *StatsService) GetMyModeratedNodes(ctx context.Context, uc *auth.UserContext) ([]model.Node, error) {
	if err := uc.RequireRole(auth.RoleModerator); err != nil {
		return nil, err
	}
	return s.GetModeratedNodesByUserID(ctx, uc.UserID)
}

// GetModeratedNodesByUserID 获取指定用户版主管辖的节点列表。
func (s *StatsService) GetModeratedNodesByUserID(ctx context.Context, userID uint) ([]model.Node, error) {
	nodeIDs, err := s.nodeModRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.log.Error("查询版主管辖节点失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	if len(nodeIDs) == 0 {
		return []model.Node{}, nil
	}
	nodes, err := s.nodeRepo.FindByIDs(ctx, nodeIDs)
	if err != nil {
		s.log.Error("查询节点信息失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return nodes, nil
}
