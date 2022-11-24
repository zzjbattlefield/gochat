package logic

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
	logic.ServiceID = "1"
	if err = logic.InitRpcServer(); err != nil {
		panic(err)
	}
}
