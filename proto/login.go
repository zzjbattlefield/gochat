package proto

type LoginResponse struct {
	AuthToken string
	Code      int
}

type LoginRequest struct {
	UserName string
	Password string
}

type RegisterRequest struct {
	UserName string
	Password string
}

type RegisterResponse struct {
	Code      int
	AuthToken string
}
