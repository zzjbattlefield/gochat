package proto

type LoginRequest struct {
	UserName string
	Password string
}

type RegisterRequest struct {
	UserName string
	Password string
}

type CheckAuthRequest struct {
	AuthToken string
}

// ------------------Reply-----------------//

type CheckAuthReponse struct {
	Code     int
	UserName string
	UserID   int
}

type RegisterResponse struct {
	Code      int
	AuthToken string
}

type LoginResponse struct {
	AuthToken string
	Code      int
}
