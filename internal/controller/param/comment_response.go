package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// CreateCommentRequest 创建评论请求。
type CreateCommentRequest struct {
	Content  string `json:"content" validate:"required,min=1"`
	ParentID *uint  `json:"parent_id"`
}

// CommentResponse 评论响应。
type CommentResponse struct {
	ID         uint              `json:"id"`
	Content    string            `json:"content"`
	User       UserResponse      `json:"user"`
	ReplyTo    *UserResponse     `json:"reply_to,omitempty"`
	LikeCount  int               `json:"like_count"`
	Liked      bool              `json:"liked"`
	ReplyCount int               `json:"reply_count"`
	Replies    []CommentResponse `json:"replies,omitempty"`
	Mentions   []MentionResponse `json:"mentions,omitempty"`
	CreatedAt  string            `json:"created_at"`
}

// CommentListResponse 评论分页列表响应。
type CommentListResponse struct {
	List     []CommentResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// ToCommentResponse 将 model.Comment 转为 API 响应。
func ToCommentResponse(c *model.Comment, likedMap map[uint]bool, mentions []model.Mention) CommentResponse {
	liked := likedMap != nil && likedMap[c.ID]
	resp := CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		User:      ToUserResponse(&c.User),
		LikeCount: c.LikeCount,
		Liked:     liked,
		CreatedAt: c.CreatedAt.Format(TimeFormat),
	}
	if c.ReplyTo != nil {
		rt := ToUserResponse(c.ReplyTo)
		resp.ReplyTo = &rt
	}
	for i := range mentions {
		resp.Mentions = append(resp.Mentions, ToMentionResponse(&mentions[i]))
	}
	return resp
}
