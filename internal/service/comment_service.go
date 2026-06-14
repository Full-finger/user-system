package service

import (
	"context"
	"errors"
	"regexp"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// urlPattern 检测评论中的 URL。
var urlPattern = regexp.MustCompile(`(?i)https?://\S+|www\.\S+`)

// CommentService 评论业务服务。
type CommentService struct {
	commentRepo     repository.CommentRepository
	commentLikeRepo repository.CommentLikeRepository
	postRepo        repository.PostRepository
	mentionRepo     repository.MentionRepository
	nodeModRepo     repository.NodeModeratorRepository
	mentionSvc      MentionParser
	txRunner        TransactionRunner
	log             *zap.Logger
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	commentLikeRepo repository.CommentLikeRepository,
	postRepo repository.PostRepository,
	mentionRepo repository.MentionRepository,
	nodeModRepo repository.NodeModeratorRepository,
	mentionSvc MentionParser,
	txRunner TransactionRunner,
	log *zap.Logger,
) *CommentService {
	return &CommentService{
		commentRepo:     commentRepo,
		commentLikeRepo: commentLikeRepo,
		postRepo:        postRepo,
		mentionRepo:     mentionRepo,
		nodeModRepo:     nodeModRepo,
		mentionSvc:      mentionSvc,
		txRunner:        txRunner,
		log:             log,
	}
}

// CreateComment 创建评论或回复。
func (s *CommentService) CreateComment(ctx context.Context, uc *auth.UserContext, postCode string, content string, parentID *uint) (*model.Comment, map[uint]bool, error) {
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return nil, nil, err
	}

	// 认证用户门槛：低于认证用户的角色不能在评论中包含链接
	if uc.Role.Level() < auth.RoleVerifiedUser.Level() && urlPattern.MatchString(content) {
		s.log.Warn("低角色用户评论包含 URL",
			zap.Uint("user_id", uc.UserID),
			zap.String("role", uc.Role.String()),
		)
		return nil, nil, apperror.BadRequest("认证用户及以上才能在评论中包含链接")
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

	if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
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
	go s.mentionSvc.ParseAndSaveCommentMentions(context.Background(), post.ID, comment.ID, content)

	// 重新加载带 User 的评论
	comment, reloadErr := s.commentRepo.FindByID(ctx, comment.ID)
	if reloadErr != nil {
		s.log.Error("创建评论后重新加载失败", zap.Uint("commentID", comment.ID), zap.Error(reloadErr))
	}
	likedMap := s.buildLikedMap(ctx, uc, []uint{comment.ID})
	return comment, likedMap, nil
}

// CommentListResult 评论列表查询结果。
type CommentListResult struct {
	Comments      []model.Comment
	Total         int64
	LikedMap      map[uint]bool
	ReplyMap      map[uint][]model.Comment
	ReplyLikedMap map[uint]map[uint]bool
	MentionMap    map[uint][]model.Mention
	ReplyCountMap map[uint]int64
}

// ListComments 获取帖子的顶级评论列表（带前 N 条回复）。
func (s *CommentService) ListComments(ctx context.Context, uc *auth.UserContext, postCode string, page, size, replyPreview int) (*CommentListResult, error) {
	post, err := s.postRepo.FindByCode(ctx, postCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("帖子不存在")
		}
		return nil, apperror.Internal("查询失败")
	}

	comments, total, err := s.commentRepo.FindByPostID(ctx, post.ID, page, size)
	if err != nil {
		s.log.Error("查询评论列表失败", zap.Error(err))
		return nil, apperror.Internal("查询失败")
	}

	if len(comments) == 0 {
		return &CommentListResult{Comments: comments, Total: total}, nil
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
	mentionMap, mentionErr := s.mentionRepo.FindByCommentIDs(ctx, allIDs)
	if mentionErr != nil {
		s.log.Warn("查询评论提及失败", zap.Error(mentionErr))
	}

	// 构建 reply likedMap
	replyLikedMap := make(map[uint]map[uint]bool)
	for parentID, rReplies := range replyMap {
		ids := commentIDs(rReplies)
		replyLikedMap[parentID] = s.buildLikedMap(ctx, uc, ids)
	}

	// 获取每个顶级评论的回复数
	parentIDs := commentIDs(comments)
	replyCountMap, countErr := s.commentRepo.CountReplies(ctx, parentIDs)
	if countErr != nil {
		s.log.Warn("查询回复数失败", zap.Error(countErr))
	}

	return &CommentListResult{
		Comments:      comments,
		Total:         total,
		LikedMap:      likedMap,
		ReplyMap:      replyMap,
		ReplyLikedMap: replyLikedMap,
		MentionMap:    mentionMap,
		ReplyCountMap: replyCountMap,
	}, nil
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
	mentionMap, mentionErr := s.mentionRepo.FindByCommentIDs(ctx, ids)
	if mentionErr != nil {
		s.log.Warn("查询回复提及失败", zap.Error(mentionErr))
	}
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
		if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
			if err := tx.Where("user_id = ? AND comment_id = ?", uc.UserID, commentID).Delete(&model.CommentLike{}).Error; err != nil {
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

	if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(&model.CommentLike{UserID: uc.UserID, CommentID: commentID}).Error; err != nil {
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

// AdminListComments 管理员/版主评论列表（支持搜索）。版主只能看管辖节点下的评论。
func (s *CommentService) AdminListComments(ctx context.Context, uc *auth.UserContext, keyword string, page, size int) ([]model.Comment, int64, error) {
	if err := uc.RequireRole(auth.RoleModerator); err != nil {
		return nil, 0, err
	}
	// 版主：只返回管辖节点下的评论
	if uc.Role == auth.RoleModerator {
		nodeIDs, err := s.nodeModRepo.FindByUserID(ctx, uc.UserID)
		if err != nil {
			s.log.Error("查询版主管辖节点失败", zap.Uint("userID", uc.UserID), zap.Error(err))
			return nil, 0, apperror.Internal("查询失败")
		}
		if len(nodeIDs) == 0 {
			return []model.Comment{}, 0, nil
		}
		comments, total, err := s.commentRepo.FindPageByNodeIDs(ctx, keyword, nodeIDs, page, size)
		if err != nil {
			s.log.Error("版主查询评论列表失败", zap.Error(err))
			return nil, 0, apperror.Internal("查询失败")
		}
		return comments, total, nil
	}
	// Admin 及以上：看全部
	comments, total, err := s.commentRepo.FindPage(ctx, keyword, page, size)
	if err != nil {
		s.log.Error("管理员查询评论列表失败", zap.Error(err))
		return nil, 0, apperror.Internal("查询失败")
	}
	return comments, total, nil
}

// AdminDeleteComment 管理员/版主删除评论。版主只能删除管辖节点下的评论。
func (s *CommentService) AdminDeleteComment(ctx context.Context, uc *auth.UserContext, id uint) error {
	if err := uc.RequireRole(auth.RoleModerator); err != nil {
		return err
	}
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("评论不存在")
		}
		return apperror.Internal("查询失败")
	}
	// 版主需检查评论所属帖子→所属节点是否在管辖范围内
	if uc.Role == auth.RoleModerator {
		post, err := s.postRepo.FindByID(ctx, comment.PostID)
		if err != nil {
			return apperror.Internal("查询失败")
		}
		ok, err := s.nodeModRepo.IsModerator(ctx, uc.UserID, post.NodeID)
		if err != nil || !ok {
			return apperror.Forbidden("无权管理该节点下的评论")
		}
	}
	if err := s.txRunner.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Comment{}, id).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ? AND reply_count > 0", comment.PostID).
			UpdateColumn("reply_count", gorm.Expr("reply_count - 1")).Error
	}); err != nil {
		s.log.Error("管理员删除评论事务失败", zap.Error(err))
		return apperror.Internal("删除评论失败")
	}
	return nil
}

// CountComments 返回评论总数。
func (s *CommentService) CountComments(ctx context.Context) (int64, error) {
	return s.commentRepo.Count(ctx)
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
