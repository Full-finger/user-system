package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// CreatePostRequest 发帖请求。
type CreatePostRequest struct {
	NodeID  uint   `json:"node_id" validate:"required"`
	Title   string `json:"title" validate:"required,min=1,max=200"`
	Content string `json:"content" validate:"required,min=1"`
}

// NodeResponse 节点简要信息。
type NodeResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	PostCount int    `json:"post_count"`
}

// MentionResponse 提及用户简要信息。
type MentionResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// PostResponse 帖子响应。
type PostResponse struct {
	ID         uint             `json:"id"`
	Title      string           `json:"title"`
	Content    string           `json:"content"`
	Node       NodeResponse     `json:"node"`
	User       UserResponse     `json:"user"`
	LikeCount  int              `json:"like_count"`
	ReplyCount int              `json:"reply_count"`
	ViewCount  int              `json:"view_count"`
	Mentions   []MentionResponse `json:"mentions,omitempty"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}

// PostListResponse 帖子分页列表响应（列表不含 mentions 以减小体积）。
type PostListResponse struct {
	List     []PostListEntry `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// PostListEntry 列表中的帖子条目。
type PostListEntry struct {
	ID         uint         `json:"id"`
	Title      string       `json:"title"`
	Content    string       `json:"content"`
	Node       NodeResponse `json:"node"`
	User       UserResponse `json:"user"`
	LikeCount  int          `json:"like_count"`
	ReplyCount int          `json:"reply_count"`
	ViewCount  int          `json:"view_count"`
	CreatedAt  string       `json:"created_at"`
	UpdatedAt  string       `json:"updated_at"`
}

// LikedPostResponse 点赞帖子响应。
type LikedPostResponse struct {
	Post PostListEntry `json:"post"`
}

// LikedPostListResponse 点赞帖子分页列表响应。
type LikedPostListResponse struct {
	List     []LikedPostResponse `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
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
		Color:     n.Color,
		Icon:      n.Icon,
		PostCount: n.PostCount,
	}
}

// ToMentionResponse 将 model.Mention 转为 API 响应。
func ToMentionResponse(m *model.Mention) MentionResponse {
	return MentionResponse{
		ID:       m.UserID,
		Username: m.Username,
	}
}

// ToPostResponse 将 model.Post + mentions 转为详情响应。
func ToPostResponse(p *model.Post, mentions []model.Mention) PostResponse {
	resp := PostResponse{
		ID:         p.ID,
		Title:      p.Title,
		Content:    p.Content,
		Node:       ToNodeResponse(&p.Node),
		User:       ToUserResponse(&p.User),
		LikeCount:  p.LikeCount,
		ReplyCount: p.ReplyCount,
		ViewCount:  p.ViewCount,
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	for i := range mentions {
		resp.Mentions = append(resp.Mentions, ToMentionResponse(&mentions[i]))
	}
	return resp
}

func toPostListEntry(p *model.Post) PostListEntry {
	return PostListEntry{
		ID:         p.ID,
		Title:      p.Title,
		Content:    p.Content,
		Node:       ToNodeResponse(&p.Node),
		User:       ToUserResponse(&p.User),
		LikeCount:  p.LikeCount,
		ReplyCount: p.ReplyCount,
		ViewCount:  p.ViewCount,
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToPostListResponse 将帖子切片转为分页列表响应。
func ToPostListResponse(posts []model.Post, total int64, page, pageSize int) PostListResponse {
	list := make([]PostListEntry, 0, len(posts))
	for i := range posts {
		list = append(list, toPostListEntry(&posts[i]))
	}
	return PostListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ToLikedPostListResponse 将点赞列表转为响应。
func ToLikedPostListResponse(likes []model.Like, total int64, page, pageSize int) LikedPostListResponse {
	list := make([]LikedPostResponse, 0, len(likes))
	for i := range likes {
		list = append(list, LikedPostResponse{
			Post: toPostListEntry(&likes[i].Post),
		})
	}
	return LikedPostListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
