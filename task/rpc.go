package task

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

var Rclient = &RpcConnectClient{
	ServiceInsMap: make(map[string][]*ClientInstance, 0),
	IndexMap:      make(map[string]int, 0),
}

type RpcConnectClient struct {
	ServiceInsMap map[string][]*ClientInstance
	IndexMap      map[string]int
	lock          sync.Mutex
}

type ClientInstance struct {
	ServiceType string
	ServiceID   string
	Client      client.XClient
}

func (task *Task) InitConnectRpcClient() (err error) {
	etcdConfig := config.Conf.Common.CommonEtcd
	etcdConfigOption := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: time.Duration(etcdConfig.ConnectionTimeout) * time.Second,
		Bucket:            "",
		PersistConnection: true,
		Username:          etcdConfig.UserName,
		Password:          etcdConfig.Password,
	}
	d, err := etcdV3.NewEtcdV3Discovery(
		etcdConfig.BasePath,
		etcdConfig.ServerPathConnect,
		[]string{etcdConfig.Host},
		true,
		etcdConfigOption,
	)
	if err != nil {
		config.Zap.Fatalf("init task rpc etcd discovery error:%v", err.Error())
	}
	if len(d.GetServices()) <= 0 {
		config.Zap.Panicf("no etcd service find")
	}
	for _, connConf := range d.GetServices() {
		config.Zap.Infof("key is %+v , value is %+v", connConf.Key, connConf.Value)
		serviceID := getParamByKey(connConf.Value, "serverId")
		serviceType := getParamByKey(connConf.Value, "serverType")
		config.Zap.Infof("serviceID is %v , serviceType is %v", serviceID, serviceType)
		if serviceID == "" || serviceType == "" {
			continue
		}
		d, err := client.NewPeer2PeerDiscovery(connConf.Key, "")
		if err != nil {
			config.Zap.Errorf("init task client.NewPeer2PeerDiscovery error :%v", err.Error())
		}
		client := client.NewXClient(etcdConfig.ServerPathConnect, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		ins := &ClientInstance{
			Client:      client,
			ServiceType: serviceType,
			ServiceID:   serviceID,
		}
		if _, ok := Rclient.ServiceInsMap[serviceID]; !ok {
			Rclient.ServiceInsMap[serviceID] = make([]*ClientInstance, 0)
		}
		Rclient.ServiceInsMap[serviceID] = append(Rclient.ServiceInsMap[serviceID], ins)
	}
	//TODO: watchServerChange
	go task.watchServerChange(d)
	return
}

func (task *Task) PushSingleToConnect(serverId string, userId int, msg []byte) {
	config.Zap.Infof("PushSingleToConnect Body %s", string(msg))
	pushMsgReq := &proto.PushRedisMessageRequest{
		UserId: userId,
		Msg: proto.Message{
			Body:      msg,
			Operation: config.OpSingleSend,
			SeqID:     tools.GetSnowFlakeId(),
		},
	}
	reply := &proto.SuccessReply{}
	client, err := Rclient.GetRpcClientByServiceID(serverId)
	if err != nil {
		config.Zap.Errorf("GetRpcClientByServiceID error:%v", err.Error())
	}
	if err = client.Call(context.Background(), "PushSingleMsg", pushMsgReq, reply); err != nil {
		config.Zap.Errorf("调用PushSingleMsg error:%v", err.Error())
	}
	config.Zap.Infof("reply : %v", reply.Msg)
}

func (task *Task) watchServerChange(d client.ServiceDiscovery) {
	etcdConfig := config.Conf.Common.CommonEtcd
	for kvChan := range d.WatchService() {
		insMap := make(map[string][]*ClientInstance)
		if len(kvChan) <= 0 {
			config.Zap.Errorf("检测到connect变更但是没有可用的ip")
		}
		config.Zap.Infoln("connect变更触发....")
		for _, kv := range kvChan {
			config.Zap.Infof("connect 变更 key:%+v, value:%+v", kv.Key, kv.Value)
			serverId := getParamByKey(kv.Value, "serverId")
			serverType := getParamByKey(kv.Value, "serverType")
			config.Zap.Infof("connect 变更 serverId:%+v , serverType:%+v", serverId, serverType)
			d, err := client.NewPeer2PeerDiscovery(kv.Key, "")
			if err != nil {
				config.Zap.Errorf("watchServerChange NewPeer2PeerDiscovery error:%v", err.Error())
			}
			c := client.NewXClient(etcdConfig.ServerPathConnect, client.Failtry, client.RandomSelect, d, client.DefaultOption)
			Instance := &ClientInstance{
				ServiceType: serverType,
				ServiceID:   serverId,
				Client:      c,
			}
			if _, ok := insMap[serverId]; !ok {
				insMap[serverId] = []*ClientInstance{Instance}
			} else {
				insMap[serverId] = append(insMap[serverId], Instance)
			}
		}
		Rclient.lock.Lock()
		Rclient.ServiceInsMap = insMap
		Rclient.lock.Unlock()
	}
}

func getParamByKey(s string, key string) string {
	param := strings.Split(s, "&")
	for _, info := range param {
		kv := strings.Split(info, "=")
		if len(kv) >= 2 && kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

func (task *Task) broadcastRoomInfoToConnect(roomID int, roomUserInfo map[string]string) {
	msg := &proto.RedisRoomUserInfo{
		RoomID:       roomID,
		Count:        len(roomUserInfo),
		RoomUserInfo: roomUserInfo,
		Op:           config.OpRoomInfoSend,
	}
	var body []byte
	var err error
	if body, err = json.Marshal(msg); err != nil {
		config.Zap.Errorf("broadcastRoomInfoToConnect json.Marshal error :%v", err.Error())
		return
	}
	req := &proto.PushRoomMessageReqeust{
		RoomID: roomID,
		Msg: proto.Message{
			Body:      body,
			Operation: config.OpRoomInfoSend,
			SeqID:     tools.GetSnowFlakeId(),
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := Rclient.GetAllConnectRpcClient()
	for _, rpc := range rpcList {
		logrus.Infof("broadcastRoomInfoToConnect rpc  %v", rpc)
		rpc.Call(context.Background(), "PushRoomInfo", req, reply)
		logrus.Infof("broadcastRoomInfoToConnect rpc  reply %v", reply)
	}
}

func (task *Task) broadcastRoomMsgToConnect(roomID int, msg []byte) {
	publishMsg := &proto.PushRoomMessageReqeust{
		RoomID: roomID,
		Msg: proto.Message{
			Body:      msg,
			Operation: config.OpRoomSend,
			SeqID:     tools.GetSnowFlakeId(),
		},
	}
	reply := &proto.SuccessReply{}
	rpcList := Rclient.GetAllConnectRpcClient()
	for _, client := range rpcList {
		client.Call(context.Background(), "PushRoomMsg", publishMsg, reply)
		config.Zap.Infof("reply %+v", reply)
	}
}

func (rc *RpcConnectClient) GetAllConnectRpcClient() (rpcClientList []client.XClient) {
	for serviceID, _ := range rc.ServiceInsMap {
		client, err := rc.GetRpcClientByServiceID(serviceID)
		if err != nil {
			config.Zap.Errorf("get client by serviceID error, service_id=%v,err is %v", serviceID, err.Error())
		}
		rpcClientList = append(rpcClientList, client)
	}
	return
}

func (rc *RpcConnectClient) GetRpcClientByServiceID(serviceID string) (c client.XClient, err error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.ServiceInsMap[serviceID]; !ok || len(rc.ServiceInsMap[serviceID]) == 0 {
		return nil, errors.New("no connect ip " + serviceID)
	}
	if _, ok := rc.IndexMap[serviceID]; !ok {
		rc.IndexMap = map[string]int{
			serviceID: 0,
		}
	}
	index := rc.IndexMap[serviceID] % len(rc.ServiceInsMap[serviceID])
	rc.IndexMap[serviceID] = (rc.IndexMap[serviceID] + 1) % len(rc.ServiceInsMap[serviceID])
	return rc.ServiceInsMap[serviceID][index].Client, nil
}
