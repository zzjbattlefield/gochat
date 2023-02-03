package logic

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
)

var RedisClient *redis.Client

func (logic *Logic) InitPublishRedisClient() (err error) {
	redisConf := config.Conf.Common.CommonRedis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisConf.RedisAddress,
		Password: redisConf.RedisPassword,
		DB:       redisConf.Db,
	})
	return nil
}

func (logic *Logic) GetRoomUserKey(RoomID string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(RoomID)
	return returnKey.String()
}

func (logic *Logic) GetUserKey(UserID string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(UserID)
	return returnKey.String()
}

func (logic *Logic) GetRoomOnlineKey(roomID string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(roomID)
	return returnKey.String()
}

func (logic *Logic) RedisPushRoomInfo(roomID int, count int, userList map[string]string) (err error) {
	redisMsg := &proto.RedisMsg{
		Op:           config.OpRoomInfoSend,
		RoomID:       roomID,
		Count:        count,
		RoomUserInfo: userList,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		config.Zap.Errorf("create json redisMsg err:%v", err.Error())
		return
	}
	config.Zap.Infof("push RoomUserInfo %+v", userList)
	if err = RedisClient.LPush(context.Background(), config.QueueName, redisMsgByte).Err(); err != nil {
		config.Zap.Errorf("push roomInfo to redis error:%v", err.Error())
	}
	return
}

func (logic *Logic) RedisPublishRoomMessage(roomID int, count int, userList map[string]string, msg []byte) (err error) {
	redisMsg := &proto.RedisMsg{
		Op:           config.OpRoomSend,
		RoomID:       roomID,
		Count:        count,
		RoomUserInfo: userList,
		Msg:          msg,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		config.Zap.Errorf("create json redisMsg err:%v", err.Error())
		return
	}
	config.Zap.Infof("publish Message %+v", string(msg))
	if err = RedisClient.LPush(context.Background(), config.QueueName, redisMsgByte).Err(); err != nil {
		config.Zap.Errorf("publish roomMsg to redis error:%v", err.Error())
	}
	return
}
