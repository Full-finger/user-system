package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// NodeResponse 节点简要信息。
type NodeResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Desc      string `json:"desc"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	PostCount int    `json:"post_count"`
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
