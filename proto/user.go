package proto

type GetUserInfoRequest struct {
	UserId int
}

type GetUserInfoResponse struct {
	UserId   int
	UserName string
	Code     int
}
