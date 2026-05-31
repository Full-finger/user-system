package param

import (
	"github.com/full-finger/user-system/internal/auth"
	"github.com/full-finger/user-system/internal/model"
)

// FollowUserResponse 关注列表中的用户简要信息。
type FollowUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// FollowResponse 关注关系响应。
type FollowResponse struct {
	ID        uint               `json:"id"`
	User      FollowUserResponse `json:"user"`
	Followed  bool               `json:"followed"`
	CreatedAt string             `json:"created_at"`
}

// FollowListResponse 关注/粉丝分页列表响应。
type FollowListResponse struct {
	List     []FollowResponse `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// UserProfileResponse 用户公开资料响应（含统计），不包含敏感字段。
type UserProfileResponse struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Nickname       string `json:"nickname"`
	Role           string `json:"role"`
	CreatedAt      string `json:"created_at"`
	PostCount      int64  `json:"post_count"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
	Followed       bool   `json:"followed"`
}

// ToFollowUserResponse 从 model.User 提取简要信息。
func ToFollowUserResponse(u *model.User) FollowUserResponse {
	return FollowUserResponse{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
	}
}

// ToFollowerListResponse 将粉丝列表转为响应。
// followedMap 可为 nil（匿名用户时所有用户 followed=false）。
func ToFollowerListResponse(follows []model.Follow, total int64, page, pageSize int, followedMap map[uint]bool) FollowListResponse {
	list := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		followed := followedMap != nil && followedMap[follows[i].FollowerID]
		list = append(list, FollowResponse{
			ID:        follows[i].ID,
			User:      ToFollowUserResponse(&follows[i].Follower),
			Followed:  followed,
			CreatedAt: follows[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
// followedMap 可为 nil（匿名用户时所有用户 followed=false）。
func ToFollowingListResponse(follows []model.Follow, total int64, page, pageSize int, followedMap map[uint]bool) FollowListResponse {
	list := make([]FollowResponse, 0, len(follows))
	for i := range follows {
		followed := followedMap != nil && followedMap[follows[i].FollowingID]
		list = append(list, FollowResponse{
			ID:        follows[i].ID,
			User:      ToFollowUserResponse(&follows[i].Following),
			Followed:  followed,
			CreatedAt: follows[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ToUserProfileResponse 将用户信息+统计转为公开资料响应。
// followed 为当前用户是否已关注该用户。
func ToUserProfileResponse(u *model.User, postCount, followerCount, followingCount int64, followed bool) UserProfileResponse {
	r := UserProfileResponse{
		ID:             u.ID,
		Username:       u.Username,
		Nickname:       u.Nickname,
		Role:           auth.Role(u.Role).String(),
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		PostCount:      postCount,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		Followed:       followed,
	}
	return r
}
