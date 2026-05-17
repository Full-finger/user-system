package model

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:user;size:20"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateRequest struct {
	Password string `json:"password" validate:"omitempty,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
