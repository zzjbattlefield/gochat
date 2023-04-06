package proto

type RedisMsg struct {
	Op           int               `json:"op"`
	Msg          []byte            `json:"msg"`
	Count        int               `json:"count"`
	ServiceIDs   []string          `json:"serviceIDs,omitempty"`
	RoomID       int               `json:"roomID,omitempty"`
	UserID       int               `json:"userID,omitempty"`
	RoomUserInfo map[string]string `json:"roomUserInfo"`
}

type RedisRoomUserInfo struct {
	RoomID       int
	RoomUserInfo map[string]string
	Count        int
	Op           int
}

type PushRedisMessageRequest struct {
	UserId int
	Msg    Message
}
