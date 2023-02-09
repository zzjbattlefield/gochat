package task

import "github.com/zzjbattlefield/IM_GO/config"

type Task struct{}

func New() *Task {
	return &Task{}
}

func (task *Task) Run() {
	if err := task.InitQueueRedisClient(); err != nil {
		config.Zap.Panicf("InitQueueRedisClient Error:%v", err.Error())
	}
	if err := task.InitConnectRpcClient(); err != nil {
		config.Zap.Panicf("InitConnectRpcClient Error:%v", err.Error())
	}
	go task.GoPush()
}
