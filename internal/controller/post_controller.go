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

// PostController 帖子相关接口的处理器。
type PostController struct {
	postSvc   *service.PostService
	nodeSvc   *service.NodeService
	followSvc *service.FollowService
	log       *zap.Logger
}

func NewPostController(postSvc *service.PostService, nodeSvc *service.NodeService, followSvc *service.FollowService, log *zap.Logger) *PostController {
	return &PostController{postSvc: postSvc, nodeSvc: nodeSvc, followSvc: followSvc, log: log}
}

// CreatePost 发帖。
func (ctrl *PostController) CreatePost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	var req param.CreatePostRequest
	if err := param.BindAndValidate(c, &req); err != nil {
		return err
	}
	post, err := ctrl.postSvc.CreatePost(c.Request().Context(), uc, req.NodeID, req.Title, req.Content)
	if err != nil {
		return err
	}
	mentions, err := ctrl.nodeSvc.GetMentions(c.Request().Context(), post.ID)
	if err != nil {
		ctrl.log.Warn("获取帖子提及列表失败", zap.Uint("postID", post.ID), zap.Error(err))
	}
	return success(c, param.ToPostResponse(post, mentions, nil))
}

// DeletePost 删帖。
func (ctrl *PostController) DeletePost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	if err := ctrl.postSvc.DeletePost(c.Request().Context(), uc, code); err != nil {
		return err
	}
	return success(c, nil)
}

// GetPost 查看帖子详情。
func (ctrl *PostController) GetPost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	post, mentions, likedMap, err := ctrl.postSvc.GetPost(c.Request().Context(), uc, code)
	if err != nil {
		return err
	}
	return success(c, param.ToPostResponse(post, mentions, likedMap))
}

// ListPosts 全站帖子列表。
func (ctrl *PostController) ListPosts(c echo.Context) error {
	uc := auth.GetUserContext(c)
	page, size := parsePage(c)
	posts, total, likedMap, err := ctrl.postSvc.ListPosts(c.Request().Context(), uc, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ListUserPosts 某用户的帖子列表。
func (ctrl *PostController) ListUserPosts(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	posts, total, likedMap, err := ctrl.postSvc.ListUserPosts(c.Request().Context(), uc, target.ID, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ListFeed 关注用户的帖子（时间线）。
func (ctrl *PostController) ListFeed(c echo.Context) error {
	uc := auth.GetUserContext(c)
	page, size := parsePage(c)
	ids, err := ctrl.followSvc.FollowingIDs(c.Request().Context(), uc)
	if err != nil {
		return err
	}
	posts, total, likedMap, err := ctrl.postSvc.ListFeed(c.Request().Context(), uc, ids, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ToggleLike 点赞/取消点赞。
func (ctrl *PostController) ToggleLike(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	liked, err := ctrl.postSvc.ToggleLike(c.Request().Context(), uc, code)
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"liked": liked})
}

// AdminListPosts 管理员帖子列表（支持搜索和节点筛选）。
func (ctrl *PostController) AdminListPosts(c echo.Context) error {
	uc := auth.GetUserContext(c)
	page, size := parsePage(c)
	keyword := c.QueryParam("keyword")
	var nodeID uint
	if nidStr := c.QueryParam("node_id"); nidStr != "" {
		if nid, err := strconv.ParseUint(nidStr, 10, 64); err == nil {
			nodeID = uint(nid)
		}
	}
	posts, total, err := ctrl.postSvc.AdminListPosts(c.Request().Context(), uc, keyword, nodeID, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToPostListResponse(posts, total, page, size, nil))
}

// AdminDeletePost 管理员/版主删帖。
func (ctrl *PostController) AdminDeletePost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("code")
	if code == "" {
		return apperror.BadRequest("无效的帖子标识")
	}
	if err := ctrl.postSvc.AdminDeletePost(c.Request().Context(), uc, code); err != nil {
		return err
	}
	return success(c, nil)
}

// ListLikedPosts 某用户点赞过的帖子。
func (ctrl *PostController) ListLikedPosts(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	likes, total, likedMap, err := ctrl.postSvc.ListLikedPosts(c.Request().Context(), uc, target.ID, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToLikedPostListResponse(likes, total, page, size, likedMap))
}
