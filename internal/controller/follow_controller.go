package controller

import (
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// FollowController 关注 / 用户资料相关接口的处理器。
type FollowController struct {
	followSvc *service.FollowService
	statsSvc  *service.StatsService
}

func NewFollowController(followSvc *service.FollowService, statsSvc *service.StatsService) *FollowController {
	return &FollowController{followSvc: followSvc, statsSvc: statsSvc}
}

// ToggleFollow 关注/取消关注。
func (ctrl *FollowController) ToggleFollow(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	followed, err := ctrl.followSvc.ToggleFollow(c.Request().Context(), uc, target.ID)
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"followed": followed})
}

// GetFollowers 某用户的粉丝列表。
func (ctrl *FollowController) GetFollowers(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, followedMap, err := ctrl.followSvc.GetFollowers(c.Request().Context(), uc, target.ID, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToFollowerListResponse(follows, total, page, size, followedMap))
}

// GetFollowings 某用户的关注列表。
func (ctrl *FollowController) GetFollowings(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, followedMap, err := ctrl.followSvc.GetFollowings(c.Request().Context(), uc, target.ID, page, size)
	if err != nil {
		return err
	}
	return success(c, param.ToFollowingListResponse(follows, total, page, size, followedMap))
}

// GetUserProfile 查看其他用户信息（含统计+版主节点）。
func (ctrl *FollowController) GetUserProfile(c echo.Context) error {
	uc := auth.GetUserContext(c)
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	// 查询用户统计
	stats, err := ctrl.statsSvc.GetUserStats(c.Request().Context(), target.ID)
	if err != nil {
		return err
	}
	// 查询关注状态
	followed, err := ctrl.followSvc.IsFollowing(c.Request().Context(), uc, target.ID)
	if err != nil {
		return err
	}
	// 查询版主管辖节点
	nodes, err := ctrl.statsSvc.GetModeratedNodesByUserID(c.Request().Context(), target.ID)
	if err != nil {
		return err
	}
	return success(c, param.ToUserProfileResponse(target, stats.PostCount, stats.FollowerCount, stats.FollowingCount, stats.LikeCount, followed, nodes))
}
