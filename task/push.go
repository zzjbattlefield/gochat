package task

import (
	"encoding/json"
	"math/rand"

	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
)

type PushParams struct {
	UserId    int
	ServiceId string
	Msg       []byte
}

var pushChannel []chan *PushParams

func init() {
	pushChannel = make([]chan *PushParams, 100)
}

func (task *Task) GoPush() {
	for i := 0; i < len(pushChannel); i++ {
		pushChannel[i] = make(chan *PushParams, 10)
		go task.processSinglePush(pushChannel[i])
	}
}

func (task *Task) processSinglePush(ch chan *PushParams) {
	for params := range ch {
		config.Zap.Infof("processSinglePush 接收到数据:%+v", string(params.Msg))
		task.PushSingleToConnect(params.ServiceId, params.UserId, params.Msg)
	}
}

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
	case config.OpSingleSend:
		pushChannel[rand.Int()%100] <- &PushParams{
			ServiceId: redisMsg.ServiceID,
			UserId:    redisMsg.UserID,
			Msg:       redisMsg.Msg,
		}
	}
}
