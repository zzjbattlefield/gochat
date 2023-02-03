package connect

import (
	"fmt"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/zzjbattlefield/IM_GO/config"
)

var DefaultService *Service

type Connect struct {
	ServiceID string
}

func New() *Connect {
	return new(Connect)
}

func (c *Connect) Run() {
	connectConfig := config.Conf.Connect
	//1.initLogicRpcClient
	c.InitLogicRpcClient()
	//2.创建buckets
	cpuNum := connectConfig.ConnectBucket.CpuNum
	//设置最大cpu数
	runtime.GOMAXPROCS(cpuNum)
	Buckets := make([]*Bucket, cpuNum)
	for i := 0; i < cpuNum; i++ {
		Buckets[i] = NewBucket(&BucketOption{
			routinueAmount: connectConfig.ConnectBucket.RoutineAmount,
		})
	}
	operator := new(DefaultOperator)
	DefaultService = NewService(Buckets, operator, ServiceOption{
		WriteWait:       10 * time.Second,
		PongWait:        60 * time.Second,
		PingPeriod:      54 * time.Second,
		MaxMessageSize:  512,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		BroadcastSize:   512,
	})
	c.ServiceID = fmt.Sprintf("ws-%s", uuid.New().String())
	if err := c.initConnectWebsocketServer(); err != nil {
		config.Zap.Errorln("initWebsocket error:", err.Error())
	}
	if err := c.initWebsocket(); err != nil {
		config.Zap.Errorln("initWebsocket error:", err.Error())
	}
}
