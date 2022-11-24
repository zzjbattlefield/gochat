package logic

import (
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
