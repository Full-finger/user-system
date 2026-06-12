package controller

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/full-finger/user-system/pkg/randstr"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CommentController 评论相关接口的处理器。
type CommentController struct {
	commentSvc *service.CommentService
	rdb        *redis.Client
	log        *zap.Logger
}

func NewCommentController(commentSvc *service.CommentService, rdb *redis.Client, log *zap.Logger) *CommentController {
	return &CommentController{commentSvc: commentSvc, rdb: rdb, log: log}
}

// GetChallenge 下发评论反垃圾 challenge（nonce + timestamp）。
func (ctrl *CommentController) GetChallenge(c echo.Context) error {
	nonce := randstr.RandomHex(16)
	ts := time.Now().UnixMilli()

	// 存储 nonce 到 Redis，TTL 25h（比 24h 上限多一些余量）
	key := fmt.Sprintf("comment:challenge:%s", nonce)
	if err := ctrl.rdb.Set(c.Request().Context(), key, "1", 25*time.Hour).Err(); err != nil {
		ctrl.log.Error("存储 challenge nonce 失败", zap.Error(err))
		return apperror.Internal("系统繁忙，请稍后重试")
	}

	return success(c, map[string]any{
		"nonce":     nonce,
		"timestamp": ts,
	})
}

// checkAntispam 执行反垃圾检查。返回 true 表示是垃圾评论（已静默处理）。
func (ctrl *CommentController) checkAntispam(c echo.Context, req *param.CreateCommentRequest) bool {
	logger := ctrl.log.With(zap.String("ip", c.RealIP()))

	// 1. 蜜罐检查：website 字段非空则判定为垃圾
	if req.Website != "" {
		logger.Warn("评论被蜜罐拦截",
			zap.String("website", req.Website),
			zap.String("content_preview", truncateStr(req.Content, 100)),
		)
		return true
	}

	// 2. 时间戳检查
	now := time.Now().UnixMilli()
	elapsed := now - req.Ts
	if elapsed < 3000 { // 小于 3 秒
		logger.Warn("评论提交过快",
			zap.Int64("elapsed_ms", elapsed),
			zap.String("content_preview", truncateStr(req.Content, 100)),
		)
		return true
	}
	if elapsed > 24*60*60*1000 { // 大于 24 小时
		logger.Warn("评论时间戳过期",
			zap.Int64("elapsed_ms", elapsed),
			zap.String("content_preview", truncateStr(req.Content, 100)),
		)
		return true
	}

	// 3. JS Challenge 检查
	if req.Nonce == "" || req.Proof == "" {
		logger.Warn("评论缺少 challenge 字段",
			zap.String("nonce", req.Nonce),
			zap.String("proof", req.Proof),
		)
		return true
	}

	// 检查 nonce 是否存在于 Redis
	key := fmt.Sprintf("comment:challenge:%s", req.Nonce)
	val, err := ctrl.rdb.Get(c.Request().Context(), key).Result()
	if err != nil || val == "" {
		logger.Warn("评论 challenge nonce 无效或已使用",
			zap.String("nonce", req.Nonce),
			zap.Error(err),
		)
		return true
	}

	// 删除 nonce（一次性使用，防重放）
	ctrl.rdb.Del(c.Request().Context(), key)

	// 验证 proof：sha256(nonce + ":" + timestamp)[:16]
	expected := computeProof(req.Nonce, req.Ts)
	if req.Proof != expected {
		logger.Warn("评论 challenge proof 不匹配",
			zap.String("nonce", req.Nonce),
			zap.String("proof", req.Proof),
			zap.String("expected", expected),
		)
		return true
	}

	return false
}

// computeProof 计算 JS Challenge 的 proof = sha256(nonce:timestamp) 的前 16 位 hex。
func computeProof(nonce string, ts int64) string {
	raw := fmt.Sprintf("%s:%d", nonce, ts)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)[:16]
}

// truncateStr 截断字符串用于日志。
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// CreateComment 创建评论或回复。
func (ctrl *CommentController) CreateComment(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	var req param.CreateCommentRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 反垃圾检查：静默丢弃（返回假的成功响应）
	if ctrl.checkAntispam(c, &req) {
		return success(c, map[string]any{
			"id":         0,
			"content":    req.Content,
			"user":       nil,
			"like_count": 0,
			"liked":      false,
			"created_at": time.Now().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	comment, likedMap, err := ctrl.commentSvc.CreateComment(c.Request().Context(), uc, code, req.Content, req.ParentID)
	if err != nil {
		return err
	}
	return success(c, param.ToCommentResponse(comment, likedMap, nil))
}

// ListComments 获取帖子的评论列表。
func (ctrl *CommentController) ListComments(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	page, size := parsePage(c)
	replyPreview, _ := strconv.Atoi(c.QueryParam("reply_preview_size"))
	if replyPreview <= 0 || replyPreview > 10 {
		replyPreview = 3
	}

	comments, total, likedMap, replyMap, replyLikedMap, mentionMap, replyCountMap, err := ctrl.commentSvc.ListComments(c.Request().Context(), uc, code, page, size, replyPreview)
	if err != nil {
		return err
	}

	list := make([]param.CommentResponse, 0, len(comments))
	for _, cm := range comments {
		mentions := mentionMap[cm.ID]
		resp := param.ToCommentResponse(&cm, likedMap, mentions)
		resp.ReplyCount = int(replyCountMap[cm.ID])

		// 填充回复预览
		if replies, ok := replyMap[cm.ID]; ok {
			resp.Replies = make([]param.CommentResponse, 0, len(replies))
			rLikedMap := replyLikedMap[cm.ID]
			for _, r := range replies {
				rMentions := mentionMap[r.ID]
				resp.Replies = append(resp.Replies, param.ToCommentResponse(&r, rLikedMap, rMentions))
			}
		}
		list = append(list, resp)
	}

	return success(c, param.CommentListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: size,
	})
}

// ListReplies 获取某评论的回复列表。
func (ctrl *CommentController) ListReplies(c echo.Context) error {
	uc := auth.GetUserContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return apperror.BadRequest("无效的评论 ID")
	}
	page, size := parsePage(c)

	replies, total, likedMap, mentionMap, err := ctrl.commentSvc.ListReplies(c.Request().Context(), uc, uint(id), page, size)
	if err != nil {
		return err
	}

	list := make([]param.CommentResponse, 0, len(replies))
	for _, r := range replies {
		mentions := mentionMap[r.ID]
		list = append(list, param.ToCommentResponse(&r, likedMap, mentions))
	}
	return success(c, param.CommentListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: size,
	})
}

// ToggleCommentLike 评论点赞/取消点赞。
func (ctrl *CommentController) ToggleCommentLike(c echo.Context) error {
	uc := auth.GetUserContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return apperror.BadRequest("无效的评论 ID")
	}
	liked, err := ctrl.commentSvc.ToggleCommentLike(c.Request().Context(), uc, uint(id))
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"liked": liked})
}

// AdminListComments 管理员评论列表（支持搜索）。
func (ctrl *CommentController) AdminListComments(c echo.Context) error {
	uc := auth.GetUserContext(c)
	page, size := parsePage(c)
	keyword := c.QueryParam("keyword")
	comments, total, err := ctrl.commentSvc.AdminListComments(c.Request().Context(), uc, keyword, page, size)
	if err != nil {
		return err
	}
	list := make([]param.CommentResponse, 0, len(comments))
	for i := range comments {
		list = append(list, param.ToCommentResponse(&comments[i], nil, nil))
	}
	return success(c, param.CommentListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: size,
	})
}

// AdminDeleteComment 管理员删除评论。
func (ctrl *CommentController) AdminDeleteComment(c echo.Context) error {
	uc := auth.GetUserContext(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return apperror.BadRequest("无效的评论 ID")
	}
	if err := ctrl.commentSvc.AdminDeleteComment(c.Request().Context(), uc, uint(id)); err != nil {
		return err
	}
	return success(c, nil)
}
