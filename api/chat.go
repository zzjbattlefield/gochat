package api

import (
	"fmt"

	"github.com/zzjbattlefield/IM_GO/api/router"
	"github.com/zzjbattlefield/IM_GO/api/rpc"
	"github.com/zzjbattlefield/IM_GO/config"
)

func Run() {
	r := router.InitRouter()
	rpc.InitLogicRpcClient()
	port := config.Conf.ApiConfig.ApiBase.ListenPort
	r.Run(fmt.Sprintf(":%d", port))
}
