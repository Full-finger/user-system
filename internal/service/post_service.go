package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/full-finger/user-system/pkg/base62"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostService 帖子业务服务。
type PostService struct {
	postRepo repository.PostRepository
	likeRepo repository.LikeRepository
	nodeRepo repository.NodeRepository
	nodeSvc  *NodeService
	log      *zap.Logger
}

func NewPostService(postRepo repository.PostRepository, likeRepo repository.LikeRepository, nodeRepo repository.NodeRepository, nodeSvc *NodeService, log *zap.Logger) *PostService {
	return &PostService{postRepo: postRepo, likeRepo: likeRepo, nodeRepo: nodeRepo, nodeSvc: nodeSvc, log: log}
}

// CreatePost 发帖。
func (s *PostService) CreatePost(ctx context.Context, userID uint, nodeID uint, title, content string) (*model.Post, error) {
	if _, err := s.nodeRepo.FindByID(ctx, nodeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.BadRequest("节点不存在")
		}
		return nil, apperror.Internal("查询节点失败")
	}

	code, err := s.generateUniqueCode(ctx)
	if err != nil {
		s.log.Error("生成帖子code失败", zap.Error(err))
		return nil, apperror.Internal("发帖失败")
	}
	post := &model.Post{
		Code:    code,
		UserID:  userID,
		NodeID:  nodeID,
		Title:   title,
		Content: content,
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		s.log.Error("发帖失败", zap.Error(err))
		return nil, apperror.Internal("发帖失败")
	}

	// 节点帖子计数 +1
	_ = s.nodeRepo.IncrPostCount(ctx, nodeID)

	// 异步解析 @提及（不阻塞主流程）
	go s.nodeSvc.ParseAndSaveMentions(ctx, post.ID, content)

	return s.postRepo.FindByID(ctx, post.ID)
}

// DeletePost 删帖，仅作者或管理员可删除。
func (s *PostService) DeletePost(ctx context.Context, userID uint, code string, isAdmin bool) error {
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("帖子不存在")
		}
		s.log.Error("查询帖子失败", zap.Error(err))
		return apperror.Internal("查询失败")
	}
	if post.UserID != userID && !isAdmin {
		return apperror.Forbidden("无权删除此帖子")
	}
	if err := s.postRepo.Delete(ctx, post.ID); err != nil {
		s.log.Error("删帖失败", zap.Error(err))
		return apperror.Internal("删帖失败")
	}
	_ = s.nodeRepo.DecrPostCount(ctx, post.NodeID)
	return nil
}

// GetPost 查看帖子详情（自增浏览数）。
func (s *PostService) GetPost(ctx context.Context, code string) (*model.Post, []model.Mention, error) {
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.NotFound("帖子不存在")
		}
		s.log.Error("查询帖子失败", zap.Error(err))
		return nil, nil, apperror.Internal("查询失败")
	}
	_ = s.postRepo.IncrViewCount(ctx, post.ID)
	post.ViewCount++

	mentions, _ := s.nodeSvc.GetMentions(ctx, post.ID)
	return post, mentions, nil
}

// ListPosts 全站帖子列表（时间倒序）。
func (s *PostService) ListPosts(ctx context.Context, page, size int) ([]model.Post, int64, error) {
	posts, total, err := s.postRepo.FindPage(ctx, page, size)
	if err != nil {
		s.log.Error("查询帖子列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return posts, total, nil
}

// ListPostsByNode 按节点查看帖子，支持 sort=time（默认）或 sort=replies。
func (s *PostService) ListPostsByNode(ctx context.Context, nodeID uint, page, size int, sort string) ([]model.Post, int64, error) {
	if _, err := s.nodeRepo.FindByID(ctx, nodeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.NotFound("节点不存在")
		}
		return nil, 0, apperror.Internal("查询失败")
	}
	posts, total, err := s.postRepo.FindByNodeID(ctx, nodeID, page, size, sort)
	if err != nil {
		s.log.Error("查询节点帖子失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return posts, total, nil
}

// ListUserPosts 某用户的帖子列表。
func (s *PostService) ListUserPosts(ctx context.Context, userID uint, page, size int) ([]model.Post, int64, error) {
	posts, total, err := s.postRepo.FindByUserID(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询用户帖子失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return posts, total, nil
}

// ListFeed 关注用户的帖子列表（时间线）。
func (s *PostService) ListFeed(ctx context.Context, followingIDs []uint, page, size int) ([]model.Post, int64, error) {
	if len(followingIDs) == 0 {
		return []model.Post{}, 0, nil
	}
	posts, total, err := s.postRepo.FindByUserIDs(ctx, followingIDs, page, size)
	if err != nil {
		s.log.Error("查询 Feed 失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return posts, total, nil
}

// ToggleLike 点赞/取消点赞，返回当前是否已赞。
func (s *PostService) ToggleLike(ctx context.Context, userID uint, code string) (bool, error) {
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperror.NotFound("帖子不存在")
		}
		return false, apperror.Internal("查询失败")
	}

	liked, err := s.likeRepo.Exists(ctx, userID, post.ID)
	if err != nil {
		s.log.Error("查询点赞状态失败", zap.Error(err))
		return false, apperror.Internal("查询失败")
	}

	if liked {
		if err := s.likeRepo.Delete(ctx, userID, post.ID); err != nil {
			s.log.Error("取消点赞失败", zap.Error(err))
			return false, apperror.Internal("取消点赞失败")
		}
		_ = s.postRepo.DecrLikeCount(ctx, post.ID)
		return false, nil
	}

	if err := s.likeRepo.Create(ctx, &model.Like{UserID: userID, PostID: post.ID}); err != nil {
		s.log.Error("点赞失败", zap.Error(err))
		return false, apperror.Internal("点赞失败")
	}
	_ = s.postRepo.IncrLikeCount(ctx, post.ID)
	return true, nil
}

// ListLikedPosts 某用户点赞过的帖子列表。
func (s *PostService) ListLikedPosts(ctx context.Context, userID uint, page, size int) ([]model.Like, int64, error) {
	likes, total, err := s.likeRepo.FindByUserID(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询点赞列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return likes, total, nil
}

// generateUniqueCode 生成唯一 base62 code，最多重试 5 次。
func (s *PostService) generateUniqueCode(ctx context.Context) (string, error) {
	const maxRetry = 5
	for i := 0; i < maxRetry; i++ {
		code := base62.Random(8)
		if _, err := s.postRepo.FindByCode(ctx, code); errors.Is(err, gorm.ErrRecordNotFound) {
			return code, nil
		}
		s.log.Warn("帖子code碰撞，重试", zap.String("code", code), zap.Int("attempt", i+1))
	}
	return "", fmt.Errorf("生成唯一code失败，已重试%d次", maxRetry)
}
