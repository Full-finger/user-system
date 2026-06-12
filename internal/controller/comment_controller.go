package controller

import (
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// CommentController 评论相关接口的处理器。
type CommentController struct {
	commentSvc *service.CommentService
	log        *zap.Logger
}

func NewCommentController(commentSvc *service.CommentService, log *zap.Logger) *CommentController {
	return &CommentController{commentSvc: commentSvc, log: log}
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
