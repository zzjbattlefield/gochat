package proto

type PushRoomMessageReqeust struct {
	Msg    string
	RoomID int
}

type Message struct {
	Body      []byte
	Operation int    //消息类型
	SeqID     string //消息唯一id
}

type ConnectRequest struct {
	AuthToken string
	RoomID    int
	ServiceID string
}

type ConnectReply struct {
	UserID int
}
