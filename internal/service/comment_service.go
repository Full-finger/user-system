package service

import (
	"context"
	"errors"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CommentService 评论业务服务。
type CommentService struct {
	commentRepo     repository.CommentRepository
	commentLikeRepo repository.CommentLikeRepository
	postRepo        repository.PostRepository
	mentionRepo     repository.MentionRepository
	nodeSvc         *NodeService
	txDB            *gorm.DB
	log             *zap.Logger
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	commentLikeRepo repository.CommentLikeRepository,
	postRepo repository.PostRepository,
	mentionRepo repository.MentionRepository,
	nodeSvc *NodeService,
	txDB *gorm.DB,
	log *zap.Logger,
) *CommentService {
	return &CommentService{
		commentRepo:     commentRepo,
		commentLikeRepo: commentLikeRepo,
		postRepo:        postRepo,
		mentionRepo:     mentionRepo,
		nodeSvc:         nodeSvc,
		txDB:            txDB,
		log:             log,
	}
}

// CreateComment 创建评论或回复。
func (s *CommentService) CreateComment(ctx context.Context, uc *auth.UserContext, postCode string, content string, parentID *uint) (*model.Comment, map[uint]bool, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, nil, err
	}

	post, err := s.postRepo.FindByCode(ctx, postCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, apperror.NotFound("帖子不存在")
		}
		return nil, nil, apperror.Internal("查询失败")
	}

	comment := &model.Comment{
		PostID:  post.ID,
		UserID:  uc.UserID,
		Content: content,
	}

	// 处理父评论（楼中楼）
	if parentID != nil {
		parent, err := s.commentRepo.FindByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, apperror.BadRequest("父评论不存在")
			}
			return nil, nil, apperror.Internal("查询失败")
		}
		if parent.PostID != post.ID {
			return nil, nil, apperror.BadRequest("父评论不属于该帖子")
		}
		comment.ParentID = parentID
		// 自动推断回复目标用户
		comment.ReplyToID = &parent.UserID
	}

	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", post.ID).
			UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).Error
	}); err != nil {
		s.log.Error("创建评论事务失败", zap.Error(err))
		return nil, nil, apperror.Internal("评论失败")
	}

	// 异步解析 @提及
	go s.nodeSvc.ParseAndSaveCommentMentions(context.Background(), post.ID, comment.ID, content)

	// 重新加载带 User 的评论
	comment, _ = s.commentRepo.FindByID(ctx, comment.ID)
	likedMap := s.buildLikedMap(ctx, uc, []uint{comment.ID})
	return comment, likedMap, nil
}

// ListComments 获取帖子的顶级评论列表（带前 N 条回复）。
func (s *CommentService) ListComments(ctx context.Context, uc *auth.UserContext, postCode string, page, size, replyPreview int) ([]model.Comment, int64, map[uint]bool, map[uint][]model.Comment, map[uint]map[uint]bool, map[uint][]model.Mention, map[uint]int64, error) {
	post, err := s.postRepo.FindByCode(ctx, postCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil, nil, nil, nil, nil, apperror.NotFound("帖子不存在")
		}
		return nil, 0, nil, nil, nil, nil, nil, apperror.Internal("查询失败")
	}

	comments, total, err := s.commentRepo.FindByPostID(ctx, post.ID, page, size)
	if err != nil {
		s.log.Error("查询评论列表失败", zap.Error(err))
		return nil, 0, nil, nil, nil, nil, nil, apperror.Internal("查询失败")
	}

	if len(comments) == 0 {
		return comments, total, nil, nil, nil, nil, nil, nil
	}

	// 收集所有评论 ID（顶级 + 回复预览）
	allIDs := commentIDs(comments)

	// 获取每个顶级评论的前 N 条回复
	replyMap := make(map[uint][]model.Comment)
	if replyPreview > 0 {
		for _, c := range comments {
			rReplies, _, err := s.commentRepo.FindReplies(ctx, c.ID, 1, replyPreview)
			if err != nil {
				s.log.Warn("查询回复预览失败", zap.Uint("commentID", c.ID), zap.Error(err))
				continue
			}
			if len(rReplies) > 0 {
				replyMap[c.ID] = rReplies
				for _, r := range rReplies {
					allIDs = append(allIDs, r.ID)
				}
			}
		}
	}

	likedMap := s.buildLikedMap(ctx, uc, allIDs)

	// 获取评论提及
	mentionMap, _ := s.mentionRepo.FindByCommentIDs(ctx, allIDs)

	// 构建 reply likedMap
	replyLikedMap := make(map[uint]map[uint]bool)
	for parentID, rReplies := range replyMap {
		ids := commentIDs(rReplies)
		replyLikedMap[parentID] = s.buildLikedMap(ctx, uc, ids)
	}

	// 获取每个顶级评论的回复数
	parentIDs := commentIDs(comments)
	replyCountMap, _ := s.commentRepo.CountReplies(ctx, parentIDs)

	return comments, total, likedMap, replyMap, replyLikedMap, mentionMap, replyCountMap, nil
}

// ListReplies 获取某评论的回复列表。
func (s *CommentService) ListReplies(ctx context.Context, uc *auth.UserContext, commentID uint, page, size int) ([]model.Comment, int64, map[uint]bool, map[uint][]model.Mention, error) {
	_, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil, nil, apperror.NotFound("评论不存在")
		}
		return nil, 0, nil, nil, apperror.Internal("查询失败")
	}

	replies, total, err := s.commentRepo.FindReplies(ctx, commentID, page, size)
	if err != nil {
		s.log.Error("查询回复列表失败", zap.Error(err))
		return nil, 0, nil, nil, apperror.Internal("查询失败")
	}

	ids := commentIDs(replies)
	likedMap := s.buildLikedMap(ctx, uc, ids)
	mentionMap, _ := s.mentionRepo.FindByCommentIDs(ctx, ids)
	return replies, total, likedMap, mentionMap, nil
}

// ToggleCommentLike 评论点赞/取消点赞。
func (s *CommentService) ToggleCommentLike(ctx context.Context, uc *auth.UserContext, commentID uint) (bool, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return false, err
	}

	_, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperror.NotFound("评论不存在")
		}
		return false, apperror.Internal("查询失败")
	}

	liked, err := s.commentLikeRepo.Exists(ctx, uc.UserID, commentID)
	if err != nil {
		s.log.Error("查询评论点赞状态失败", zap.Error(err))
		return false, apperror.Internal("查询失败")
	}

	if liked {
		if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := s.commentLikeRepo.Delete(ctx, uc.UserID, commentID); err != nil {
				return err
			}
			return tx.Model(&model.Comment{}).Where("id = ? AND like_count > 0", commentID).
				UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
		}); err != nil {
			s.log.Error("取消评论点赞事务失败", zap.Error(err))
			return false, apperror.Internal("取消点赞失败")
		}
		return false, nil
	}

	if err := s.txDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.commentLikeRepo.Create(ctx, &model.CommentLike{UserID: uc.UserID, CommentID: commentID}); err != nil {
			return err
		}
		return tx.Model(&model.Comment{}).Where("id = ?", commentID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	}); err != nil {
		s.log.Error("评论点赞事务失败", zap.Error(err))
		return false, apperror.Internal("点赞失败")
	}
	return true, nil
}

// GetCommentMentions 获取评论的提及列表。
func (s *CommentService) GetCommentMentions(ctx context.Context, commentIDs []uint) (map[uint][]model.Mention, error) {
	return s.mentionRepo.FindByCommentIDs(ctx, commentIDs)
}

func (s *CommentService) buildLikedMap(ctx context.Context, uc *auth.UserContext, ids []uint) map[uint]bool {
	if uc.IsGuest() || len(ids) == 0 {
		return nil
	}
	m, err := s.commentLikeRepo.FindLikedCommentIDs(ctx, uc.UserID, ids)
	if err != nil {
		s.log.Warn("查询评论点赞状态失败", zap.Error(err))
	}
	return m
}

// commentIDs 从评论切片提取 ID 列表。
func commentIDs(comments []model.Comment) []uint {
	ids := make([]uint, 0, len(comments))
	for i := range comments {
		ids = append(ids, comments[i].ID)
	}
	return ids
}
