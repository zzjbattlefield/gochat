package proto

type PushRoomMessageReqeust struct {
	Msg    Message
	RoomID int
}

type Message struct {
	Body      []byte
	Operation int    //消息类型
	SeqID     string //消息唯一id
}

type ConnectRequest struct {
	AuthToken string `json:"authToken"`
	RoomID    int    `json:"roomId"`
	ServiceID string `json:"serverId"`
}

type DisConnectRequest struct {
	RoomID int
	UserID int
}

type DisConnectReply struct {
	Has bool
}

type ConnectReply struct {
	UserID int
}
