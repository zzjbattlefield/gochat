package tools

import (
	"sync"

	"github.com/go-redis/redis/v8"
)

var RedisClientMap = map[string]*redis.Client{}
var syncLock sync.Mutex

type RedisOpt struct {
	Address  string
	Password string
	DB       int
}

// 获取redis实例
func GetRedisInstance(opt *RedisOpt) (client *redis.Client) {
	syncLock.Lock()
	defer syncLock.Unlock()
	if client, ok := RedisClientMap[opt.Address]; ok {
		return client
	}
	client = redis.NewClient(&redis.Options{
		Addr:     opt.Address,
		Password: opt.Password,
		DB:       opt.DB,
	})
	RedisClientMap[opt.Address] = client
	return
}
