package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// MentionResponse 提及用户简要信息。
type MentionResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// PostResponse 帖子响应。
type PostResponse struct {
	Code       string            `json:"code"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Node       NodeResponse      `json:"node"`
	User       UserResponse      `json:"user"`
	LikeCount  int               `json:"like_count"`
	ReplyCount int               `json:"reply_count"`
	ViewCount  int               `json:"view_count"`
	Liked      bool              `json:"liked"`
	Mentions   []MentionResponse `json:"mentions,omitempty"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
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
	Code       string       `json:"code"`
	Title      string       `json:"title"`
	Content    string       `json:"content"`
	Node       NodeResponse `json:"node"`
	User       UserResponse `json:"user"`
	LikeCount  int          `json:"like_count"`
	ReplyCount int          `json:"reply_count"`
	ViewCount  int          `json:"view_count"`
	Liked      bool         `json:"liked"`
	CreatedAt  string       `json:"created_at"`
	UpdatedAt  string       `json:"updated_at"`
}

// LikedPostResponse 点赞帖子响应。
type LikedPostResponse struct {
	Post  PostListEntry `json:"post"`
	Liked bool          `json:"liked"`
}

// LikedPostListResponse 点赞帖子分页列表响应。
type LikedPostListResponse struct {
	List     []LikedPostResponse `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// ToMentionResponse 将 model.Mention 转为 API 响应。
func ToMentionResponse(m *model.Mention) MentionResponse {
	return MentionResponse{
		ID:       m.UserID,
		Username: m.Username,
	}
}

// ToPostResponse 将 model.Post + mentions + likedMap 转为详情响应。
// likedMap 可为 nil（匿名用户时所有帖子 liked=false）。
func ToPostResponse(p *model.Post, mentions []model.Mention, likedMap map[uint]bool) PostResponse {
	liked := likedMap != nil && likedMap[p.ID]
	resp := PostResponse{
		Code:       p.Code,
		Title:      p.Title,
		Content:    p.Content,
		Node:       ToNodeResponse(&p.Node),
		User:       ToUserResponse(&p.User),
		LikeCount:  p.LikeCount,
		ReplyCount: p.ReplyCount,
		ViewCount:  p.ViewCount,
		Liked:      liked,
		CreatedAt:  p.CreatedAt.Format(TimeFormat),
		UpdatedAt:  p.UpdatedAt.Format(TimeFormat),
	}
	for i := range mentions {
		resp.Mentions = append(resp.Mentions, ToMentionResponse(&mentions[i]))
	}
	return resp
}

func toPostListEntry(p *model.Post, likedMap map[uint]bool) PostListEntry {
	liked := likedMap != nil && likedMap[p.ID]
	return PostListEntry{
		Code:       p.Code,
		Title:      p.Title,
		Content:    truncate(p.Content, 200),
		Node:       ToNodeResponse(&p.Node),
		User:       ToUserResponse(&p.User),
		LikeCount:  p.LikeCount,
		ReplyCount: p.ReplyCount,
		ViewCount:  p.ViewCount,
		Liked:      liked,
		CreatedAt:  p.CreatedAt.Format(TimeFormat),
		UpdatedAt:  p.UpdatedAt.Format(TimeFormat),
	}
}

// ToPostListResponse 将帖子切片转为分页列表响应。
// likedMap 可为 nil（匿名用户时所有帖子 liked=false）。
func ToPostListResponse(posts []model.Post, total int64, page, pageSize int, likedMap map[uint]bool) PostListResponse {
	list := make([]PostListEntry, 0, len(posts))
	for i := range posts {
		list = append(list, toPostListEntry(&posts[i], likedMap))
	}
	return PostListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ToLikedPostListResponse 将点赞列表转为响应。
// likedMap 可为 nil（游客）；Liked 表示当前查看者是否也点赞了该帖子。
func ToLikedPostListResponse(likes []model.Like, total int64, page, pageSize int, likedMap map[uint]bool) LikedPostListResponse {
	list := make([]LikedPostResponse, 0, len(likes))
	for i := range likes {
		// Liked 表示当前查看者对该帖子是否也点了赞（非目标用户的点赞状态）
		liked := likedMap != nil && likedMap[likes[i].PostID]
		list = append(list, LikedPostResponse{
			Post:  toPostListEntry(&likes[i].Post, likedMap),
			Liked: liked,
		})
	}
	return LikedPostListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// truncate 将字符串硬截断为 n 个字符（rune），超出部分用 "..." 替代。
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
