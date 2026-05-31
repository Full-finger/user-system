package controller

import (
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// parsePage 从查询参数解析分页信息，默认 page=1, size=20。
func parsePage(c echo.Context) (page, size int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	size, _ = strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return
}

// resolveUsername 从路由参数中解析用户名并查找用户，失败时返回错误响应。
func resolveUsername(c echo.Context, followSvc *service.FollowService) (*model.User, error) {
	username := c.Param("username")
	if username == "" {
		return nil, apperror.BadRequest("缺少用户名")
	}
	return followSvc.ResolveUsername(c.Request().Context(), username)
}

// buildLikedMap 从帖子列表构建 likedMap，Guest 时返回 nil。
func buildLikedMap(c echo.Context, posts []model.Post, likeSvc *service.LikeService) map[uint]bool {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].ID)
	}
	return buildLikedMapForPosts(c, ids, likeSvc)
}

// buildLikedMapForPosts 从帖子 ID 列表构建 likedMap，Guest 时返回 nil。
func buildLikedMapForPosts(c echo.Context, postIDs []uint, likeSvc *service.LikeService) map[uint]bool {
	uc := auth.GetUserContext(c)
	if uc.IsGuest() || len(postIDs) == 0 {
		return nil
	}
	m, _ := likeSvc.FindLikedPostIDs(c.Request().Context(), uc, postIDs)
	return m
}

// buildFollowedMap 从用户 ID 列表构建 followedMap，Guest 时返回 nil。
func buildFollowedMap(c echo.Context, userIDs []uint, followSvc *service.FollowService) map[uint]bool {
	uc := auth.GetUserContext(c)
	if uc.IsGuest() || len(userIDs) == 0 {
		return nil
	}
	m, _ := followSvc.FindFollowedUserIDs(c.Request().Context(), uc, userIDs)
	return m
}
