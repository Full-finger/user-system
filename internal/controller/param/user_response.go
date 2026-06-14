package param

import (
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/avatar"
	"github.com/full-finger/user-system/internal/model"
)

// UserResponse 用户信息响应，脱敏密码。
type UserResponse struct {
	ID             uint            `json:"id"`
	Username       string          `json:"username"`
	Nickname       string          `json:"nickname"`
	Email          string          `json:"email,omitempty"`
	AvatarURL      string          `json:"avatar_url"`
	CoverTheme     string          `json:"cover_theme"`
	Motto          string          `json:"motto"`
	Role           string          `json:"role"`
	CreatedAt      string          `json:"created_at"`
	PostCount      int64           `json:"post_count,omitempty"`
	CommentCount   int64           `json:"comment_count,omitempty"`
	LikeCount      int64           `json:"like_count,omitempty"`
	FollowerCount  int64           `json:"follower_count,omitempty"`
	FollowingCount int64           `json:"following_count,omitempty"`
	LikedCount     int64           `json:"liked_count,omitempty"`
	ModeratedNodes []ModeratedNode `json:"moderated_nodes,omitempty"`
}

// ModeratedNode 版主管辖的节点简要信息。
type ModeratedNode struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UserListResponse 分页用户列表响应。
type UserListResponse struct {
	List     []UserResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ToUserResponse 将 model.User 转为 API 响应。
func ToUserResponse(u *model.User) UserResponse {
	r := UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		AvatarURL: avatar.AvatarURL(u.Email, u.Username),
		Role:      auth.Role(u.Role).String(),
		CreatedAt: u.CreatedAt.Format(TimeFormat),
	}
	if u.Email != nil {
		r.Email = *u.Email
	}
	if u.CoverTheme != nil {
		r.CoverTheme = *u.CoverTheme
	}
	if u.Motto != nil {
		r.Motto = *u.Motto
	}
	return r
}

// ToUserResponseWithStats 将用户信息+统计数据转为 API 响应。
func ToUserResponseWithStats(u *model.User, postCount, commentCount, likeCount, followerCount, followingCount, likedCount int64) UserResponse {
	r := ToUserResponse(u)
	r.PostCount = postCount
	r.CommentCount = commentCount
	r.LikeCount = likeCount
	r.FollowerCount = followerCount
	r.FollowingCount = followingCount
	r.LikedCount = likedCount
	return r
}

// ToUserResponseWithModeratedNodes 将用户信息+管辖节点转为 API 响应。
func ToUserResponseWithModeratedNodes(u *model.User, nodes []model.Node) UserResponse {
	r := ToUserResponse(u)
	if len(nodes) > 0 {
		r.ModeratedNodes = make([]ModeratedNode, len(nodes))
		for i, n := range nodes {
			r.ModeratedNodes[i] = ModeratedNode{ID: n.ID, Name: n.Name, Slug: n.Slug}
		}
	}
	return r
}

// ToUserListResponse 将用户切片转为分页列表响应。
func ToUserListResponse(users []model.User, total int64, page, pageSize int) UserListResponse {
	list := make([]UserResponse, 0, len(users))
	for i := range users {
		list = append(list, ToUserResponse(&users[i]))
	}
	return UserListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
