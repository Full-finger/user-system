package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/full-finger/user-system/pkg/base62"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostService 帖子业务服务。
type PostService struct {
	postRepo    repository.PostRepository
	likeRepo    repository.LikeRepository
	likeSvc     *LikeService
	nodeRepo    repository.NodeRepository
	nodeModRepo repository.NodeModeratorRepository
	nodeSvc     *NodeService
	txDB        *gorm.DB
	log         *zap.Logger
}

func NewPostService(postRepo repository.PostRepository, likeRepo repository.LikeRepository, likeSvc *LikeService, nodeRepo repository.NodeRepository, nodeModRepo repository.NodeModeratorRepository, nodeSvc *NodeService, txDB *gorm.DB, log *zap.Logger) *PostService {
	return &PostService{postRepo: postRepo, likeRepo: likeRepo, likeSvc: likeSvc, nodeRepo: nodeRepo, nodeModRepo: nodeModRepo, nodeSvc: nodeSvc, txDB: txDB, log: log}
}

// CreatePost 发帖。
func (s *PostService) CreatePost(ctx context.Context, uc *auth.UserContext, nodeID uint, title, content string) (*model.Post, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, err
	}
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
		UserID:  uc.UserID,
		NodeID:  nodeID,
		Title:   title,
		Content: content,
	}

	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Node{}).Where("id = ?", nodeID).
			UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.log.Error("发帖事务失败", zap.Error(err))
		return nil, apperror.Internal("发帖失败")
	}

	go s.nodeSvc.ParseAndSaveMentions(context.Background(), post.ID, content)
	return s.postRepo.FindByID(ctx, post.ID)
}

// DeletePost 删帖：作者本人 / 版主（管辖该节点）/ 管理员及以上。
func (s *PostService) DeletePost(ctx context.Context, uc *auth.UserContext, code string) error {
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("帖子不存在")
		}
		s.log.Error("查询帖子失败", zap.Error(err))
		return apperror.Internal("查询失败")
	}

	if !s.canDeletePost(ctx, uc, post) {
		return apperror.Forbidden("无权删除此帖子")
	}

	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Post{}, post.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Node{}).Where("id = ? AND post_count > 0", post.NodeID).
			UpdateColumn("post_count", gorm.Expr("post_count - 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.log.Error("删帖事务失败", zap.Error(err))
		return apperror.Internal("删帖失败")
	}
	return nil
}

func (s *PostService) canDeletePost(ctx context.Context, uc *auth.UserContext, post *model.Post) bool {
	// 作者本人
	if uc.UserID == post.UserID {
		return true
	}
	// Admin 及以上
	if uc.Role.Level() >= auth.RoleAdmin.Level() {
		return true
	}
	// Moderator 检查节点管辖
	if uc.Role == auth.RoleModerator {
		ok, err := s.nodeModRepo.IsModerator(ctx, uc.UserID, post.NodeID)
		return err == nil && ok
	}
	return false
}

// GetPost 查看帖子详情（自增浏览数）。
func (s *PostService) GetPost(ctx context.Context, uc *auth.UserContext, code string) (*model.Post, []model.Mention, map[uint]bool, error) {
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, apperror.NotFound("帖子不存在")
		}
		s.log.Error("查询帖子失败", zap.Error(err))
		return nil, nil, nil, apperror.Internal("查询失败")
	}
	_ = s.postRepo.IncrViewCount(ctx, post.ID)
	post.ViewCount++

	mentions, _ := s.nodeSvc.GetMentions(ctx, post.ID)
	likedMap := s.buildLikedMap(ctx, uc, []uint{post.ID})
	return post, mentions, likedMap, nil
}

// ListPosts 全站帖子列表（时间倒序）。
func (s *PostService) ListPosts(ctx context.Context, uc *auth.UserContext, page, size int) ([]model.Post, int64, map[uint]bool, error) {
	posts, total, err := s.postRepo.FindPage(ctx, page, size)
	if err != nil {
		s.log.Error("查询帖子列表失败", zap.Error(err))
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	likedMap := s.buildLikedMap(ctx, uc, postIDs(posts))
	return posts, total, likedMap, nil
}

// ListPostsByNode 按节点查看帖子。
func (s *PostService) ListPostsByNode(ctx context.Context, uc *auth.UserContext, nodeID uint, page, size int, sort string) ([]model.Post, int64, map[uint]bool, error) {
	if _, err := s.nodeRepo.FindByID(ctx, nodeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil, apperror.NotFound("节点不存在")
		}
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	posts, total, err := s.postRepo.FindByNodeID(ctx, nodeID, page, size, sort)
	if err != nil {
		s.log.Error("查询节点帖子失败", zap.Error(err))
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	likedMap := s.buildLikedMap(ctx, uc, postIDs(posts))
	return posts, total, likedMap, nil
}

// ListUserPosts 某用户的帖子列表。
func (s *PostService) ListUserPosts(ctx context.Context, uc *auth.UserContext, userID uint, page, size int) ([]model.Post, int64, map[uint]bool, error) {
	posts, total, err := s.postRepo.FindByUserID(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询用户帖子失败", zap.Error(err))
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	likedMap := s.buildLikedMap(ctx, uc, postIDs(posts))
	return posts, total, likedMap, nil
}

// ListFeed 关注用户的帖子列表（时间线）。
func (s *PostService) ListFeed(ctx context.Context, uc *auth.UserContext, followingIDs []uint, page, size int) ([]model.Post, int64, map[uint]bool, error) {
	if len(followingIDs) == 0 {
		return []model.Post{}, 0, nil, nil
	}
	posts, total, err := s.postRepo.FindByUserIDs(ctx, followingIDs, page, size)
	if err != nil {
		s.log.Error("查询 Feed 失败", zap.Error(err))
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	likedMap := s.buildLikedMap(ctx, uc, postIDs(posts))
	return posts, total, likedMap, nil
}

// ToggleLike 点赞/取消点赞，返回当前是否已赞。
func (s *PostService) ToggleLike(ctx context.Context, uc *auth.UserContext, code string) (bool, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return false, err
	}
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperror.NotFound("帖子不存在")
		}
		return false, apperror.Internal("查询失败")
	}

	liked, err := s.likeRepo.Exists(ctx, uc.UserID, post.ID)
	if err != nil {
		s.log.Error("查询点赞状态失败", zap.Error(err))
		return false, apperror.Internal("查询失败")
	}

	if liked {
		if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND post_id = ?", uc.UserID, post.ID).Delete(&model.Like{}).Error; err != nil {
				return err
			}
			return tx.Model(&model.Post{}).Where("id = ? AND like_count > 0", post.ID).
				UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
		}); err != nil {
			s.log.Error("取消点赞事务失败", zap.Error(err))
			return false, apperror.Internal("取消点赞失败")
		}
		return false, nil
	}

	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Like{UserID: uc.UserID, PostID: post.ID}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", post.ID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	}); err != nil {
		s.log.Error("点赞事务失败", zap.Error(err))
		return false, apperror.Internal("点赞失败")
	}
	return true, nil
}

// AdminDeletePost 管理员/版主删帖：Moderator 及以上，版主需管辖该节点。
func (s *PostService) AdminDeletePost(ctx context.Context, uc *auth.UserContext, code string) error {
	if err := uc.RequireRole(auth.RoleModerator); err != nil {
		return err
	}
	post, err := s.postRepo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("帖子不存在")
		}
		s.log.Error("查询帖子失败", zap.Error(err))
		return apperror.Internal("查询失败")
	}
	// 版主需检查节点管辖
	if uc.Role == auth.RoleModerator {
		ok, err := s.nodeModRepo.IsModerator(ctx, uc.UserID, post.NodeID)
		if err != nil || !ok {
			return apperror.Forbidden("无权管理该节点的帖子")
		}
	}
	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Post{}, post.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Node{}).Where("id = ? AND post_count > 0", post.NodeID).
			UpdateColumn("post_count", gorm.Expr("post_count - 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.log.Error("管理员删帖事务失败", zap.Error(err))
		return apperror.Internal("删帖失败")
	}
	return nil
}

// ListLikedPosts 某用户点赞过的帖子列表。
func (s *PostService) ListLikedPosts(ctx context.Context, uc *auth.UserContext, userID uint, page, size int) ([]model.Like, int64, map[uint]bool, error) {
	likes, total, err := s.likeRepo.FindByUserID(ctx, userID, page, size)
	if err != nil {
		s.log.Error("查询点赞列表失败", zap.Error(err))
		return nil, 0, nil, apperror.Internal("查询失败")
	}
	ids := make([]uint, 0, len(likes))
	for i := range likes {
		ids = append(ids, likes[i].PostID)
	}
	likedMap := s.buildLikedMap(ctx, uc, ids)
	return likes, total, likedMap, nil
}

// buildLikedMap 根据用户身份构建点赞映射，Guest 返回 nil。
func (s *PostService) buildLikedMap(ctx context.Context, uc *auth.UserContext, ids []uint) map[uint]bool {
	if uc.IsGuest() || len(ids) == 0 {
		return nil
	}
	m, _ := s.likeSvc.FindLikedPostIDs(ctx, uc, ids)
	return m
}

// postIDs 从帖子切片提取 ID 列表。
func postIDs(posts []model.Post) []uint {
	ids := make([]uint, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].ID)
	}
	return ids
}

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
