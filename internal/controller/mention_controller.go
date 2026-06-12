package controller

import (
	"strconv"
	"strings"

	"github.com/full-finger/user-system/internal/apperror"
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/controller/param"
	"github.com/full-finger/user-system/internal/repository"
	"github.com/labstack/echo/v4"
)

// MentionController @提及补全相关接口。
type MentionController struct {
	userRepo    repository.UserRepository
	followRepo  repository.FollowRepository
	nodeModRepo repository.NodeModeratorRepository
}

func NewMentionController(userRepo repository.UserRepository, followRepo repository.FollowRepository, nodeModRepo repository.NodeModeratorRepository) *MentionController {
	return &MentionController{userRepo: userRepo, followRepo: followRepo, nodeModRepo: nodeModRepo}
}

// MentionSuggestionUser @补全候选用户简要信息。
type MentionSuggestionUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// GetMentionCache 获取当前用户可用于 @补全的用户列表。
// GET /api/mention-cache?sources=following,followers,admins,moderators&node_id=1
func (ctrl *MentionController) GetMentionCache(c echo.Context) error {
	uc := auth.GetUserContext(c)
	if err := uc.RequireRole(auth.RoleUser); err != nil {
		return err
	}

	sourcesStr := c.QueryParam("sources")
	if sourcesStr == "" {
		sourcesStr = "following,followers,admins,moderators"
	}
	sources := parseSources(sourcesStr)

	ctx := c.Request().Context()
	seen := make(map[uint]struct{})
	var allIDs []uint

	for _, src := range sources {
		var ids []uint
		switch src {
		case "following":
			followingIDs, err := ctrl.followRepo.FollowingIDs(ctx, uc.UserID)
			if err != nil {
				return apperror.Internal("查询失败")
			}
			ids = followingIDs
		case "followers":
			followerIDs, err := ctrl.followRepo.FollowerIDs(ctx, uc.UserID)
			if err != nil {
				return apperror.Internal("查询失败")
			}
			ids = followerIDs
		case "admins":
			admins, err := ctrl.userRepo.FindByRoleGTE(ctx, int(auth.RoleAdmin))
			if err != nil {
				return apperror.Internal("查询失败")
			}
			for _, a := range admins {
				ids = append(ids, a.ID)
			}
		case "moderators":
			nodeIDStr := c.QueryParam("node_id")
			if nodeIDStr != "" {
				nodeID, err := strconv.ParseUint(nodeIDStr, 10, 64)
				if err == nil {
					modIDs, _ := ctrl.nodeModRepo.FindUserIDsByNodeID(ctx, uint(nodeID))
					ids = modIDs
				}
			}
		}
		for _, id := range ids {
			if id == uc.UserID {
				continue // 不包括自己
			}
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				allIDs = append(allIDs, id)
			}
		}
	}

	if len(allIDs) == 0 {
		return success(c, []MentionSuggestionUser{})
	}

	users, err := ctrl.userRepo.FindByIDs(ctx, allIDs)
	if err != nil {
		return apperror.Internal("查询失败")
	}

	// 保持 allIDs 的顺序（去重后的合并顺序）
	orderMap := make(map[uint]int, len(allIDs))
	for i, id := range allIDs {
		orderMap[id] = i
	}
	// 按 orderMap 排序
	sorted := make([]MentionSuggestionUser, len(allIDs))
	for _, u := range users {
		idx, ok := orderMap[u.ID]
		if ok {
			nickname := u.Nickname
			if nickname == "" {
				nickname = u.Username
			}
			sorted[idx] = MentionSuggestionUser{
				ID:       u.ID,
				Username: u.Username,
				Nickname: nickname,
			}
		}
	}
	// 过滤零值（理论上不会出现）
	result := make([]MentionSuggestionUser, 0, len(sorted))
	for _, s := range sorted {
		if s.ID != 0 {
			result = append(result, s)
		}
	}

	return success(c, result)
}

func parseSources(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Ensure param package can reference MentionSuggestionUser if needed.
// We also add a helper for comment controller to check user existence.
var _ = param.MentionResponse{}
