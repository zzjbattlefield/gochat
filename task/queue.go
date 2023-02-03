package task

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/tools"
)

var RedisClient *redis.Client

func (task *Task) InitQueueRedisClient() (err error) {
	taskConf := config.Conf.TaskConfig.TaskBase
	redisOpt := tools.RedisOpt{
		Address:  taskConf.RedisAddress,
		Password: taskConf.RedisPassword,
		DB:       0,
	}
	RedisClient = tools.GetRedisInstance(&redisOpt)
	if pong, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		config.Zap.Infof("RedisClient Ping Result Pong:%v, err:%v", pong, err.Error())
	}
	go func() {
		for {
			result, err := RedisClient.BRPop(context.Background(), time.Second*10, config.QueueName).Result()
			if err != nil {
				config.Zap.Infof("Redis BRPop error:%v", err.Error())
			}
			if len(result) >= 2 {
				task.Push(result[1])
			}
		}
	}()
	return
}
