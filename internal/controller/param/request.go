package param

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
