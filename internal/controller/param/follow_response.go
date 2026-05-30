package param

import (
	"github.com/full-finger/user-system/internal/model"
)

// FollowUserResponse 关注列表中的用户简要信息。
type FollowUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// FollowResponse 关注关系响应。
type FollowResponse struct {
	ID        uint               `json:"id"`
	User      FollowUserResponse `json:"user"`
	CreatedAt string             `json:"created_at"`
}

// FollowListResponse 关注/粉丝分页列表响应。
type FollowListResponse struct {
	List     []FollowResponse `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// UserProfileResponse 用户详情响应（含统计）。
type UserProfileResponse struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email,omitempty"`
	Role           string `json:"role"`
	CreatedAt      string `json:"created_at"`
	PostCount      int64  `json:"post_count"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
}

// ToFollowUserResponse 从 model.User 提取简要信息。
func ToFollowUserResponse(u *model.User) FollowUserResponse {
	return FollowUserResponse{
		ID:       u.ID,
		Username: u.Username,
	}
}

// ToFollowerListResponse 将粉丝列表转为响应。
func ToFollowerListResponse(follows []model.Follow, total int64, page, pageSize int) FollowListResponse {
	list := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		list = append(list, FollowResponse{
			ID:        follows[i].ID,
			User:      ToFollowUserResponse(&follows[i].Follower),
			CreatedAt: follows[i].Follower.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ToFollowingListResponse 将关注列表转为响应。
func ToFollowingListResponse(follows []model.Follow, total int64, page, pageSize int) FollowListResponse {
	list := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		list = append(list, FollowResponse{
			ID:        follows[i].ID,
			User:      ToFollowUserResponse(&follows[i].Following),
			CreatedAt: follows[i].Following.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ToUserProfileResponse 将用户信息+统计转为响应。
func ToUserProfileResponse(u *model.User, postCount, followerCount, followingCount int64) UserProfileResponse {
	r := UserProfileResponse{
		ID:             u.ID,
		Username:       u.Username,
		Role:           u.Role,
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		PostCount:      postCount,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
	}
	if u.Email != nil {
		r.Email = *u.Email
	}
	return r
}
