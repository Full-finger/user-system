package service

import (
	"context"
	"errors"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FollowService 关注业务服务。
type FollowService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
	postRepo   repository.PostRepository
	log        *zap.Logger
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository, postRepo repository.PostRepository, log *zap.Logger) *FollowService {
	return &FollowService{followRepo: followRepo, userRepo: userRepo, postRepo: postRepo, log: log}
}

// ToggleFollow 关注/取消关注，返回当前是否已关注。
func (s *FollowService) ToggleFollow(ctx context.Context, followerID, followingID uint) (bool, error) {
	if followerID == followingID {
		return false, apperror.BadRequest("不能关注自己")
	}
	if _, err := s.userRepo.FindByID(ctx, followingID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return false, apperror.Internal("查询失败")
	}

	followed, err := s.followRepo.Exists(ctx, followerID, followingID)
	if err != nil {
		s.log.Error("查询关注状态失败", zap.Error(err))
		return false, apperror.Internal("查询失败")
	}

	if followed {
		if err := s.followRepo.Delete(ctx, followerID, followingID); err != nil {
			s.log.Error("取消关注失败", zap.Error(err))
			return false, apperror.Internal("取消关注失败")
		}
		return false, nil
	}

	if err := s.followRepo.Create(ctx, &model.Follow{FollowerID: followerID, FollowingID: followingID}); err != nil {
		s.log.Error("关注失败", zap.Error(err))
		return false, apperror.Internal("关注失败")
	}
	return true, nil
}

// GetFollowers 获取某用户的粉丝列表。
func (s *FollowService) GetFollowers(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error) {
	follows, total, err := s.followRepo.FindFollowers(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询粉丝列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return follows, total, nil
}

// GetFollowings 获取某用户的关注列表。
func (s *FollowService) GetFollowings(ctx context.Context, userID uint, page, size int) ([]model.Follow, int64, error) {
	follows, total, err := s.followRepo.FindFollowings(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询关注列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return follows, total, nil
}

// GetUserStats 获取用户的统计信息（帖子数、粉丝数、关注数）。
func (s *FollowService) GetUserStats(ctx context.Context, userID uint) (postCount int64, followerCount int64, followingCount int64, err error) {
	_, postCount, err = s.postRepo.FindByUserID(ctx, userID, 1, 1)
	if err != nil {
		s.log.Error("查询用户帖子数失败", zap.Error(err))
		return 0, 0, 0, apperror.Internal("查询失败")
	}
	followerCount, err = s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		s.log.Error("查询粉丝数失败", zap.Error(err))
		return 0, 0, 0, apperror.Internal("查询失败")
	}
	followingCount, err = s.followRepo.CountFollowings(ctx, userID)
	if err != nil {
		s.log.Error("查询关注数失败", zap.Error(err))
		return 0, 0, 0, apperror.Internal("查询失败")
	}
	return postCount, followerCount, followingCount, nil
}

// GetUserProfile 获取用户资料+统计。
func (s *FollowService) GetUserProfile(ctx context.Context, userID uint) (*model.User, int64, int64, int64, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, 0, 0, apperror.NotFound("用户不存在")
		}
		s.log.Error("查询用户失败", zap.Error(err))
		return nil, 0, 0, 0, apperror.Internal("查询失败")
	}
	postCount, followerCount, followingCount, err := s.GetUserStats(ctx, userID)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return user, postCount, followerCount, followingCount, nil
}

// FollowingIDs 获取当前用户关注的所有用户ID。
func (s *FollowService) FollowingIDs(ctx context.Context, userID uint) ([]uint, error) {
	ids, err := s.followRepo.FollowingIDs(ctx, userID)
	if err != nil {
		s.log.Error("查询关注列表失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return ids, nil
}

// IsFollowing 判断 followerID 是否已关注 followingID。
func (s *FollowService) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	return s.followRepo.Exists(ctx, followerID, followingID)
}

// FindFollowedUserIDs 批量查询当前用户关注了哪些用户，返回 userID → bool 映射。
func (s *FollowService) FindFollowedUserIDs(ctx context.Context, followerID uint, userIDs []uint) (map[uint]bool, error) {
	m, err := s.followRepo.FindFollowedUserIDs(ctx, followerID, userIDs)
	if err != nil {
		s.log.Error("批量查询关注状态失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}
	return m, nil
}
