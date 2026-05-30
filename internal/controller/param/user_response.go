package param

import "github.com/full-finger/user-system/internal/model"

// UserResponse 用户信息响应，脱敏密码。
type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
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
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.Email != nil {
		r.Email = *u.Email
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
