package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// ModeratorInfo 版主简要信息。
type ModeratorInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// NodeResponse 节点简要信息。
type NodeResponse struct {
	ID         uint            `json:"id"`
	Name       string          `json:"name"`
	Slug       string          `json:"slug"`
	Desc       string          `json:"desc"`
	Color      string          `json:"color"`
	Icon       string          `json:"icon"`
	PostCount  int             `json:"post_count"`
	Moderators []ModeratorInfo `json:"moderators,omitempty"`
}

// NodeListResponse 节点列表响应。
type NodeListResponse struct {
	Nodes []NodeResponse `json:"nodes"`
}

// ToNodeResponse 将 model.Node 转为 API 响应。
func ToNodeResponse(n *model.Node) NodeResponse {
	return NodeResponse{
		ID:        n.ID,
		Name:      n.Name,
		Slug:      n.Slug,
		Desc:      n.Desc,
		Color:     n.Color,
		Icon:      n.Icon,
		PostCount: n.PostCount,
	}
}

// ToNodeResponseWithModerators 将节点+版主列表转为 API 响应。
func ToNodeResponseWithModerators(n *model.Node, mods []ModeratorInfo) NodeResponse {
	r := ToNodeResponse(n)
	r.Moderators = mods
	return r
}
