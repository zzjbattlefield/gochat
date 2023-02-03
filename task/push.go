package task

import (
	"encoding/json"

	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
)

func (task *Task) Push(msg string) {
	redisMsg := &proto.RedisMsg{}
	if err := json.Unmarshal([]byte(msg), redisMsg); err != nil {
		config.Zap.Infof("Json Unmarshal Error:%v", err.Error())
	}
	config.Zap.Infof("Push Msg Info %+v,Op is %d", msg, redisMsg.Op)
	switch redisMsg.Op {
	case config.OpRoomInfoSend:
		task.broadcastRoomInfoToConnect(redisMsg.RoomID, redisMsg.RoomUserInfo)
	case config.OpRoomSend:
		task.broadcastRoomMsgToConnect(redisMsg.RoomID, redisMsg.Msg)
	}
}
