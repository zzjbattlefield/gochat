package rpc

import (
	"context"
	"sync"

	"github.com/smallnest/rpcx/client"
	"github.com/zzjbattlefield/IM_GO/proto"
)

var LogicRpcClient client.XClient
var once sync.Once

type RpcLogic struct{}

var RpcLoginObj *RpcLogic

// 初始化对logicRpc的客户端初始化
func InitLogicRpcClient() {
	once.Do(func() {
		d, err := client.NewPeer2PeerDiscovery("tcp@127.0.0.1:6900", "")
		if err != nil {
			panic(err)
		}
		LogicRpcClient = client.NewXClient("LogicRpc", client.Failtry, client.RandomSelect, d, client.DefaultOption)

		RpcLoginObj = new(RpcLogic)
	})
	if LogicRpcClient == nil {
		panic("rpc client启动失败")
	}
}

func (rpc *RpcLogic) Login(request *proto.LoginRequest) (code int, authToken string, msg string) {
	reply := new(proto.LoginResponse)
	err := LogicRpcClient.Call(context.Background(), "Login", request, reply)
	if err != nil {
		msg = err.Error()
	}
	authToken = reply.AuthToken
	code = reply.Code
	return
}

func (rpc *RpcLogic) Register(request *proto.RegisterRequest) (code int, authToken string, msg string) {
	reply := new(proto.RegisterResponse)
	err := LogicRpcClient.Call(context.Background(), "Register", request, reply)
	if err != nil {
		msg = err.Error()
	}
	authToken = reply.AuthToken
	code = reply.Code
	return
}
