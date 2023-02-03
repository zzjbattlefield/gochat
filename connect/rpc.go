package connect

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

var LogicRpcClient client.XClient
var once sync.Once

type RpcConnect struct {
}

func (c *Connect) InitLogicRpcClient() (err error) {
	once.Do(func() {
		d, err := client.NewPeer2PeerDiscovery("tcp@127.0.0.1:6900", "")
		if err != nil {
			panic(err)
		}
		LogicRpcClient = client.NewXClient("LogicRpc", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	})
	if LogicRpcClient == nil {
		panic("rpc client启动失败")
	}
	return nil
}

func (c *RpcConnect) Connect(connReq *proto.ConnectRequest) (userID int, err error) {
	reply := &proto.ConnectReply{}
	if err = LogicRpcClient.Call(context.Background(), "Connect", connReq, reply); err != nil {
		config.Zap.Errorln("fail to call Connect:", err)
	}
	config.Zap.Infoln("get connect info userID:", reply.UserID)
	userID = reply.UserID
	return
}

func (c *RpcConnect) DisConnect(req *proto.DisConnectRequest) error {
	reply := &proto.DisConnectReply{}
	return LogicRpcClient.Call(context.Background(), "DisConnect", req, reply)
}

func (c *Connect) initConnectWebsocketServer() (err error) {
	var network, addr string
	connectRpcAddress := strings.Split(config.Conf.Connect.ConnectRpcAddressWebSocket.Address, ",")
	for _, bind := range connectRpcAddress {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			config.Zap.Panicf("InitConnectWebsocketRpcServer ParseNetwork error : %s", err)
		}
		config.Zap.Infof("Connect start run at-->%s:%s", network, addr)
		go c.createConnectWebsocktsRpcServer(network, addr)
	}
	return
}

func (c *Connect) createConnectWebsocktsRpcServer(network string, addr string) {
	s := server.NewServer()
	addRegistryPlugin(s, network, addr)
	config.Zap.Infoln("network & addr :", network, " , ", addr)
	config.Zap.Infof("ServerPathConnect:%+v", config.Conf.Common.CommonEtcd.ServerPathConnect)
	s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathConnect, new(RpcConnectPush), fmt.Sprintf("serverId=%s&serverType=ws", c.ServiceID))
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}

type RpcConnectPush struct {
}

func (rpc *RpcConnectPush) PushRoomInfo(ctx context.Context, pushRoomMsg *proto.PushRoomMessageReqeust, successReply *proto.SuccessReply) (err error) {
	successReply.Code = config.SuccessReplyCode
	successReply.Msg = "success"
	config.Zap.Infof("pushRoomMsg msg:%+v", pushRoomMsg)
	for _, bucket := range DefaultService.Buckets {
		bucket.BroadcastRoom(pushRoomMsg)
	}
	return
}

func (rpc *RpcConnectPush) PushRoomMsg(ctx context.Context, msg *proto.PushRoomMessageReqeust, reply *proto.SuccessReply) (err error) {
	reply.Code = config.SuccessReplyCode
	reply.Msg = "success"
	for _, bucket := range DefaultService.Buckets {
		bucket.BroadcastRoom(msg)
	}
	return
}

func addRegistryPlugin(s *server.Server, network, address string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + address,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	config.Zap.Infof("etcdConfig:%+v", r)
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
