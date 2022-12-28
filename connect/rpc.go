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

type ConnectRpc struct {
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

func (c *ConnectRpc) Connect(connReq *proto.ConnectRequest) (userID int, err error) {
	reply := &proto.ConnectReply{}
	if err = LogicRpcClient.Call(context.Background(), "Connect", connReq, reply); err != nil {
		config.Zap.Errorln("fail to call Connect:", err)
	}
	config.Zap.Infoln("get connect info userID:", reply.UserID)
	userID = reply.UserID
	return
}

func (c *ConnectRpc) DisConnect(req *proto.DisConnectRequest) error {
	reply := &proto.DisConnectReply{}
	return LogicRpcClient.Call(context.Background(), "DisConnect", req, reply)
}

func (c *Connect) initConnectWebsocketServer() (err error) {
	var network, address string
	connConfig := config.Conf.Connect
	connectRpcAddress := strings.Split(connConfig.ConnectRpcAddressWebSocket.Address, ",")
	for _, bind := range connectRpcAddress {
		network, address, err = tools.ParseNetwork(bind)
		if err != nil {
			config.Zap.Errorf("初始化connect rpcx server错误 %s", err.Error())
			return
		}
		go func(network, address string) {
			s := server.NewServer()
			addRegistryPlugin(s, network, address)
			s.RegisterName(config.Conf.Common.CommentEtcd.ServerPathConnect, new(ConnectRpc), fmt.Sprintf("serviceID=%s&serviceType=ws", c.ServiceID))
			s.Serve(network, address)
		}(network, address)
	}
	return
}
func addRegistryPlugin(s *server.Server, network, address string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + address,
		EtcdServers:    []string{config.Conf.Common.CommentEtcd.Host},
		BasePath:       config.Conf.Common.CommentEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
