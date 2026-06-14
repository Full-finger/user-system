package controller

import (
	"strings"

	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/service"
	"github.com/labstack/echo/v4"
)

// MentionController @提及补全相关接口。
type MentionController struct {
	mentionSvc *service.MentionService
}

func NewMentionController(mentionSvc *service.MentionService) *MentionController {
	return &MentionController{mentionSvc: mentionSvc}
}

// mentionSuggestionResponse @补全候选用户响应。
type mentionSuggestionResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// GetMentionCache 获取当前用户可用于 @补全的用户列表。
// GET /api/mention-cache?sources=following,followers,admins,moderators&node_id=1
func (ctrl *MentionController) GetMentionCache(c echo.Context) error {
	uc := auth.GetUserContext(c)

	sourcesStr := c.QueryParam("sources")
	if sourcesStr == "" {
		sourcesStr = "following,followers,admins,moderators"
	}
	sources := parseSources(sourcesStr)
	nodeIDStr := c.QueryParam("node_id")

	users, err := ctrl.mentionSvc.GetMentionCache(c.Request().Context(), uc, sources, nodeIDStr)
	if err != nil {
		return err
	}

	result := make([]mentionSuggestionResponse, len(users))
	for i, u := range users {
		result[i] = mentionSuggestionResponse{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
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
