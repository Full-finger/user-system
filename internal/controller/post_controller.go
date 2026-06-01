package controller

import (
	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// PostController 帖子相关接口的处理器。
type PostController struct {
	postSvc   *service.PostService
	nodeSvc   *service.NodeService
	followSvc *service.FollowService
	likeSvc   *service.LikeService
}

func NewPostController(postSvc *service.PostService, nodeSvc *service.NodeService, followSvc *service.FollowService, likeSvc *service.LikeService) *PostController {
	return &PostController{postSvc: postSvc, nodeSvc: nodeSvc, followSvc: followSvc, likeSvc: likeSvc}
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
	mentions, _ := ctrl.nodeSvc.GetMentions(c.Request().Context(), post.ID)
	return success(c, param.ToPostResponse(post, mentions, nil))
}

// DeletePost 删帖。
func (ctrl *PostController) DeletePost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("id")
	if code == "" {
		return apperror.BadRequest("无效的ID")
	}
	if err := ctrl.postSvc.DeletePost(c.Request().Context(), uc, code); err != nil {
		return err
	}
	return success(c, nil)
}

// GetPost 查看帖子详情。
func (ctrl *PostController) GetPost(c echo.Context) error {
	code := c.Param("id")
	if code == "" {
		return apperror.BadRequest("无效的ID")
	}
	post, mentions, err := ctrl.postSvc.GetPost(c.Request().Context(), code)
	if err != nil {
		return err
	}
	likedMap := buildLikedMapForPosts(c, []uint{post.ID}, ctrl.likeSvc)
	return success(c, param.ToPostResponse(post, mentions, likedMap))
}

// ListPosts 全站帖子列表。
func (ctrl *PostController) ListPosts(c echo.Context) error {
	page, size := parsePage(c)
	posts, total, err := ctrl.postSvc.ListPosts(c.Request().Context(), page, size)
	if err != nil {
		return err
	}
	likedMap := buildLikedMap(c, posts, ctrl.likeSvc)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ListUserPosts 某用户的帖子列表。
func (ctrl *PostController) ListUserPosts(c echo.Context) error {
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	posts, total, err := ctrl.postSvc.ListUserPosts(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	likedMap := buildLikedMap(c, posts, ctrl.likeSvc)
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
	posts, total, err := ctrl.postSvc.ListFeed(c.Request().Context(), ids, page, size)
	if err != nil {
		return err
	}
	likedMap := buildLikedMap(c, posts, ctrl.likeSvc)
	return success(c, param.ToPostListResponse(posts, total, page, size, likedMap))
}

// ToggleLike 点赞/取消点赞。
func (ctrl *PostController) ToggleLike(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("id")
	if code == "" {
		return apperror.BadRequest("无效的ID")
	}
	liked, err := ctrl.postSvc.ToggleLike(c.Request().Context(), uc, code)
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"liked": liked})
}

// AdminDeletePost 管理员/版主删帖。
func (ctrl *PostController) AdminDeletePost(c echo.Context) error {
	uc := auth.GetUserContext(c)
	code := c.Param("id")
	if code == "" {
		return apperror.BadRequest("无效的ID")
	}
	if err := ctrl.postSvc.AdminDeletePost(c.Request().Context(), uc, code); err != nil {
		return err
	}
	return success(c, nil)
}

// ListLikedPosts 某用户点赞过的帖子。
func (ctrl *PostController) ListLikedPosts(c echo.Context) error {
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	likes, total, err := ctrl.postSvc.ListLikedPosts(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	postIDs := make([]uint, 0, len(likes))
	for i := range likes {
		postIDs = append(postIDs, likes[i].PostID)
	}
	likedMap := buildLikedMapForPosts(c, postIDs, ctrl.likeSvc)
	return success(c, param.ToLikedPostListResponse(likes, total, page, size, likedMap))
}
