package param

import "github.com/full-finger/user-system/internal/model"

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type UserListResponse struct {
	List     []UserResponse `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

func ToUserResponse(u *model.User) UserResponse {
	r := UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.Email != nil {
		r.Email = *u.Email
	}
	return r
}

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
