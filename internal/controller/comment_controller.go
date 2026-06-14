package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/config"
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
	antispam   config.AntispamConfig
}

func NewCommentController(commentSvc *service.CommentService, rdb *redis.Client, log *zap.Logger, antispam config.AntispamConfig) *CommentController {
	return &CommentController{commentSvc: commentSvc, rdb: rdb, log: log, antispam: antispam.Defaults()}
}

// GetChallenge 下发评论反垃圾 challenge（nonce + timestamp + difficulty + honeypot_field）。
func (ctrl *CommentController) GetChallenge(c echo.Context) error {
	// 频率限制：每分钟最多 ChallengeLimit 次
	ip := c.RealIP()
	limitKey := fmt.Sprintf("antispam:challenge_limit:%s", ip)
	count, err := ctrl.rdb.Incr(c.Request().Context(), limitKey).Result()
	if err == nil {
		if count == 1 {
			ctrl.rdb.Expire(c.Request().Context(), limitKey, time.Minute)
		}
		if count > int64(ctrl.antispam.ChallengeLimit) {
			return apperror.TooMany("请求过于频繁，请稍后再试")
		}
	}

	nonce := randstr.RandomHex(16)
	ts := time.Now().UnixMilli()

	// 存储 nonce + timestamp 到 Redis，TTL 25h（比 24h 上限多一些余量）
	key := fmt.Sprintf("comment:challenge:%s", nonce)
	val := strconv.FormatInt(ts, 10)
	if err := ctrl.rdb.Set(c.Request().Context(), key, val, 25*time.Hour).Err(); err != nil {
		ctrl.log.Error("存储 challenge nonce 失败", zap.Error(err))
		return apperror.Internal("系统繁忙，请稍后重试")
	}

	return success(c, map[string]any{
		"nonce":          nonce,
		"timestamp":      ts,
		"difficulty":     ctrl.antispam.Difficulty,
		"honeypot_field": ctrl.antispam.HoneypotField,
	})
}

// CreateComment 创建评论或回复。
func (ctrl *CommentController) CreateComment(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}

	// 读取原始 body 用于提取动态蜜罐字段
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return apperror.BadRequest("参数错误")
	}
	// 恢复 body 供后续 Bind 使用
	c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// 从原始 JSON 中按配置字段名提取蜜罐值
	honeypotValue := ""
	var raw map[string]any
	if json.Unmarshal(bodyBytes, &raw) == nil {
		if v, ok := raw[ctrl.antispam.HoneypotField]; ok {
			if s, ok := v.(string); ok {
				honeypotValue = s
			}
		}
	}

	var req param.CreateCommentRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 反 bot 检查：静默丢弃（蜜罐用动态提取的值）
	if isSpam, isBot := ctrl.checkAntispamWithHoneypot(c, &req, honeypotValue); isSpam {
		if isBot {
			return success(c, map[string]any{
				"id":         0,
				"content":    req.Content,
				"user":       nil,
				"like_count": 0,
				"liked":      false,
				"created_at": time.Now().Format(param.TimeFormat),
			})
		}
	}

	// 内容过滤检查：给用户反馈
	if msg := ctrl.checkContentFilter(c, &req); msg != "" {
		return apperror.BadRequest(msg)
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

	result, err := ctrl.commentSvc.ListComments(c.Request().Context(), uc, code, page, size, replyPreview)
	if err != nil {
		return err
	}

	list := make([]param.CommentResponse, 0, len(result.Comments))
	for _, cm := range result.Comments {
		mentions := result.MentionMap[cm.ID]
		resp := param.ToCommentResponse(&cm, result.LikedMap, mentions)
		resp.ReplyCount = int(result.ReplyCountMap[cm.ID])

		if replies, ok := result.ReplyMap[cm.ID]; ok {
			resp.Replies = make([]param.CommentResponse, 0, len(replies))
			rLikedMap := result.ReplyLikedMap[cm.ID]
			for _, r := range replies {
				rMentions := result.MentionMap[r.ID]
				resp.Replies = append(resp.Replies, param.ToCommentResponse(&r, rLikedMap, rMentions))
			}
		}
		list = append(list, resp)
	}

	return success(c, param.CommentListResponse{
		List:     list,
		Total:    result.Total,
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
