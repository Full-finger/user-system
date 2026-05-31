package controller

import (
	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// FollowController 关注 / 用户资料相关接口的处理器。
type FollowController struct {
	followSvc *service.FollowService
	likeSvc   *service.LikeService
}

func NewFollowController(followSvc *service.FollowService, likeSvc *service.LikeService) *FollowController {
	return &FollowController{followSvc: followSvc, likeSvc: likeSvc}
}

// ToggleFollow 关注/取消关注。
func (ctrl *FollowController) ToggleFollow(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return apperror.Unauthorized("未认证")
	}
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	followed, err := ctrl.followSvc.ToggleFollow(c.Request().Context(), userID, target.ID)
	if err != nil {
		return err
	}
	return success(c, map[string]bool{"followed": followed})
}

// GetFollowers 某用户的粉丝列表。
func (ctrl *FollowController) GetFollowers(c echo.Context) error {
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, err := ctrl.followSvc.GetFollowers(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	userIDs := make([]uint, 0, len(follows))
	for i := range follows {
		userIDs = append(userIDs, follows[i].FollowerID)
	}
	followedMap := buildFollowedMap(c, userIDs, ctrl.followSvc)
	return success(c, param.ToFollowerListResponse(follows, total, page, size, followedMap))
}

// GetFollowings 某用户的关注列表。
func (ctrl *FollowController) GetFollowings(c echo.Context) error {
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	page, size := parsePage(c)
	follows, total, err := ctrl.followSvc.GetFollowings(c.Request().Context(), target.ID, page, size)
	if err != nil {
		return err
	}
	userIDs := make([]uint, 0, len(follows))
	for i := range follows {
		userIDs = append(userIDs, follows[i].FollowingID)
	}
	followedMap := buildFollowedMap(c, userIDs, ctrl.followSvc)
	return success(c, param.ToFollowingListResponse(follows, total, page, size, followedMap))
}

// GetUserProfile 查看其他用户信息（含统计）。
func (ctrl *FollowController) GetUserProfile(c echo.Context) error {
	target, err := resolveUsername(c, ctrl.followSvc)
	if err != nil {
		return err
	}
	user, postCount, followerCount, followingCount, err := ctrl.followSvc.GetUserProfile(c.Request().Context(), target.ID)
	if err != nil {
		return err
	}
	followed := false
	if currentID, ok := optionalUserID(c); ok && currentID != target.ID {
		followed, _ = ctrl.followSvc.IsFollowing(c.Request().Context(), currentID, target.ID)
	}
	return success(c, param.ToUserProfileResponse(user, postCount, followerCount, followingCount, followed))
}
