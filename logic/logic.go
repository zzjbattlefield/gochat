package logic

import "github.com/google/uuid"

type Logic struct {
	ServiceID string
}

func New() *Logic {
	return &Logic{}
}

func (logic *Logic) Run() {
	err := logic.InitPublishRedisClient()
	if err != nil {
		panic(err)
	}
	logic.ServiceID = uuid.New().String()
	if err = logic.InitRpcServer(); err != nil {
		panic(err)
	}
}
