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

type Send struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	FromUserId   int    `json:"fromUserId"`
	FromUserName string `json:"fromUserName"`
	ToUserId     int    `json:"toUserId"`
	ToUserName   string `json:"toUserName"`
	RoomId       int    `json:"roomId"`
	Op           int    `json:"op"`
	CreateTime   string `json:"createTime"`
}

type SendTcp struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	FromUserId   int    `json:"fromUserId"`
	FromUserName string `json:"fromUserName"`
	ToUserId     int    `json:"toUserId"`
	ToUserName   string `json:"toUserName"`
	RoomId       int    `json:"roomId"`
	Op           int    `json:"op"`
	CreateTime   string `json:"createTime"`
	AuthToken    string `json:"authToken"`
}

type SuccessReply struct {
	Code int
	Msg  string
}
