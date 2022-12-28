package logic

import (
	"bytes"

	"github.com/go-redis/redis/v8"
	"github.com/zzjbattlefield/IM_GO/config"
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
