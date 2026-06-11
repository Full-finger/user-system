package controller

import (
	"strconv"

	"github.com/full-finger/user-system/internal/apperror"
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
