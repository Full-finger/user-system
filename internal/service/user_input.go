package service

type RegisterInput struct {
	Username string
	Password string
	Email    string
}

type LoginInput struct {
	Username string
	Password string
}

type UpdateInput struct {
	Password string
	Role     string
}
