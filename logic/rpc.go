package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/model"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

type LogicRpc struct{}

func (logic *Logic) InitRpcServer() (err error) {
	var (
		network string
		address string
	)
	binds := strings.Split(config.Conf.LogicConfig.LogicBase.RpcAddress, ",")
	for _, bind := range binds {
		if network, address, err = tools.ParseNetwork(bind); err != nil {
			config.Zap.Fatalf("InitRpcServer ParseNetwork err:%s", err.Error())
			return
		}
		go logic.createLogicRpcServer(network, address)
	}
	return
}

func (logic *Logic) createLogicRpcServer(network, address string) {
	s := server.NewServer()
	addRegistryPlugin(s, network, address)
	metadata := fmt.Sprintf("logic-%s", logic.ServiceID)
	s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(LogicRpc), metadata)
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	if err := s.Serve(network, address); err != nil {
		config.Zap.Fatalf("createLogicRpcServer rpc Serve err:%s", err.Error())
	}
}

func addRegistryPlugin(s *server.Server, network, address string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + address,
		EtcdServers:    []string{config.Conf.Common.CommonEtcd.Host},
		BasePath:       config.Conf.Common.CommonEtcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	if err := r.Start(); err != nil {
		config.Zap.Fatalf("addRegistryPlugin err:%s", err.Error())
	}
	s.Plugins.Add(r)
}

func (rpc *LogicRpc) Register(ctx context.Context, request *proto.RegisterRequest, reply *proto.RegisterResponse) (err error) {
	reply.Code = config.FailReplyCode
	model := new(model.UserModel)
	data := model.CheckHaveUserName(request.UserName)
	if data.ID != 0 {
		return errors.New("用户已经存在 请登录")
	}
	model.Password = tools.Md5(request.Password)
	model.UserName = request.UserName
	userID, err := model.Add()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if userID == 0 {
		return errors.New("新增用户失败")
	}
	//构建token
	sessionID, err := CreateAuthToken(ctx, model)
	if err != nil {
		return err
	}
	reply.Code = config.SuccessReplyCode
	reply.AuthToken = sessionID
	return
}

func (rpc *LogicRpc) Login(ctx context.Context, request *proto.LoginRequest, reply *proto.LoginResponse) (err error) {
	reply.Code = tools.CodeFail
	userName := request.UserName
	password := tools.Md5(request.Password)
	model := new(model.UserModel)
	userInfo := model.CheckHaveUserName(userName)
	if userInfo.ID == 0 || userInfo.Password != password {
		return errors.New("用户名或密码错误")
	}
	sessionID, err := CreateAuthToken(ctx, userInfo)
	if err != nil {
		return err
	}
	reply.Code = tools.CodeSuccess
	reply.AuthToken = sessionID
	return
}

func (rpc *LogicRpc) Connect(ctx context.Context, request *proto.ConnectRequest, reply *proto.ConnectReply) (err error) {
	logic := new(Logic)
	config.Zap.Infoln("get args authToken is:", request.AuthToken)
	sessionID := tools.GetSessionName(request.AuthToken)
	userInfo, err := RedisClient.HGetAll(ctx, sessionID).Result()
	if err != nil {
		config.Zap.Errorf("redis HGetAll Key: %s error: %s", sessionID, err.Error())
		return err
	}
	reply.UserID, _ = strconv.Atoi(userInfo["userID"])
	if len(userInfo) <= 0 {
		reply.UserID = 0
		return
	}
	roomUserkey := logic.GetRoomUserKey(strconv.Itoa(request.RoomID))
	userKey := logic.GetUserKey(userInfo["userID"])

	//绑定当前用户所在的serviceID
	// err = RedisClient.Set(ctx, userKey, request.ServiceID, config.RedisBaseValidTime*time.Second).Err()
	//使用redis集合保存用户所在的serviceID
	err = RedisClient.SAdd(ctx, userKey, request.ServiceID).Err()
	if err != nil {
		config.Zap.Errorf("redis SAdd error: %s", err.Error())
		return
	}
	if RedisClient.HGet(ctx, roomUserkey, userInfo["userID"]).Val() == "" {
		RedisClient.HSet(ctx, roomUserkey, userInfo["userID"], userInfo["userName"])
		RedisClient.Incr(ctx, logic.GetRoomOnlineKey(strconv.Itoa(request.RoomID)))
	}
	config.Zap.Infoln("logic rpc userID", reply.UserID)
	return
}

func (rpc *LogicRpc) DisConnect(ctx context.Context, request *proto.DisConnectRequest, reply *proto.DisConnectReply) (err error) {
	logic := new(Logic)
	roomUserKey := logic.GetRoomUserKey(strconv.Itoa(request.RoomID))
	count, _ := RedisClient.Get(ctx, logic.GetRoomOnlineKey(strconv.Itoa(request.RoomID))).Int()
	if count > 0 {
		RedisClient.Decr(ctx, logic.GetRoomOnlineKey(strconv.Itoa(request.RoomID))).Result()
	}
	if request.UserID > 0 {
		if err = RedisClient.HDel(ctx, roomUserKey, strconv.Itoa(request.UserID)).Err(); err != nil {
			config.Zap.Warnf("RedisCli HGetAll roomUserInfo key:%s, err: %s", roomUserKey, err)
		}
		//广播一下当前的房间信息
		userList, err := RedisClient.HGetAll(ctx, roomUserKey).Result()
		if err != nil {
			config.Zap.Errorf("Disconnect Get UserList Error:%v", err.Error())
			return err
		}
		err = logic.RedisPushRoomInfo(request.RoomID, count-1, userList)
		if err != nil {
			config.Zap.Errorf("Disconnect Send RoomInfo Error:%v", err.Error())
			return err
		}
	}
	return
}

func (rpc *LogicRpc) CheckAuth(ctx context.Context, request *proto.CheckAuthRequest, reply *proto.CheckAuthReponse) (err error) {
	var tokenVal = make(map[string]string)
	reply.Code = tools.CodeFail
	authToken := request.AuthToken
	tokenVal, err = RedisClient.HGetAll(ctx, tools.CreateSessionId(authToken)).Result()
	if err != nil || len(tokenVal) == 0 {
		config.Zap.Errorw("检测authToken失败", "authToken", authToken)
		return
	}
	reply.UserID, _ = strconv.Atoi(tokenVal["userID"])
	reply.UserName = tokenVal["userName"]
	reply.Code = tools.CodeSuccess
	return
}

func CreateAuthToken(ctx context.Context, user *model.UserModel) (randStr string, err error) {
	randStr = tools.GetRandString(32)
	sessionID := tools.CreateSessionId(randStr)
	sessionData := make(map[string]interface{})
	sessionData["userName"] = user.UserName
	sessionData["userID"] = user.ID
	err = RedisClient.HMSet(ctx, sessionID, sessionData).Err()
	if err != nil {
		return
	}
	err = RedisClient.Expire(ctx, sessionID, 86400*time.Second).Err()
	if err != nil {
		return
	}
	return
}

func (rpc *LogicRpc) GetRoomInfo(ctx context.Context, request *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = tools.CodeFail
	logic := new(Logic)
	roomID := request.RoomId
	roomUserList := make(map[string]string)
	roomUserKey := logic.GetRoomUserKey(strconv.Itoa(roomID))
	roomUserList, err = RedisClient.HGetAll(context.Background(), roomUserKey).Result()
	if err != nil {
		config.Zap.Errorf("redis get roomInfo err:%v", err.Error())
		return
	}
	if len(roomUserList) == 0 {
		return fmt.Errorf("get no user list from room:%d", roomID)
	}
	if err = logic.RedisPushRoomInfo(roomID, len(roomUserList), roomUserList); err == nil {
		reply.Code = tools.CodeSuccess
	}
	return
}

func (rpc *LogicRpc) PushRoom(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	reply.Code = tools.CodeFail
	logic := new(Logic)
	roomId := args.RoomId
	roomUserInfo := make(map[string]string)
	userKey := logic.GetRoomUserKey(strconv.Itoa(roomId))
	roomUserInfo, err = RedisClient.HGetAll(ctx, userKey).Result()
	if err != nil {
		config.Zap.Errorf("logic redis获取userKey错误 userKey:%+v, err:%v", userKey, err.Error())
		return
	}
	args.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	var bodyBytes []byte
	bodyBytes, err = json.Marshal(args)
	if err != nil {
		config.Zap.Errorf("logic jsonMarshal 错误:%v", err.Error())
		return
	}
	err = logic.RedisPublishRoomMessage(roomId, len(roomUserInfo), roomUserInfo, bodyBytes)
	reply.Code = tools.CodeSuccess
	return
}

func (rpc *LogicRpc) GetUserInfoByUserId(ctx context.Context, args *proto.GetUserInfoRequest, reply *proto.GetUserInfoResponse) (err error) {
	userId := args.UserId
	reply.Code = config.FailReplyCode
	userModel := new(model.UserModel)
	if err = userModel.GetUserInfoByUserId(userId); err != nil {
		config.Zap.Errorf("获取用户信息失败:%v", err.Error())
		return
	}
	reply.Code = config.SuccessReplyCode
	reply.UserId = userModel.ID
	reply.UserName = userModel.UserName
	return
}

func (rpc *LogicRpc) Push(ctx context.Context, args *proto.Send, reply *proto.SuccessReply) (err error) {
	var (
		bodyByte   []byte
		serviceIds []string
	)
	reply.Code = config.FailReplyCode
	bodyByte, err = json.Marshal(args)
	if err != nil {
		config.Zap.Errorf("json.Marshal 错误:%v", err)
		return
	}
	logic := new(Logic)
	userKey := logic.GetUserKey(strconv.Itoa(args.ToUserId))
	if serviceIds, err = RedisClient.SMembers(ctx, userKey).Result(); err != nil {
		config.Zap.Errorf("redis获取user下的serverID错误 userKey:%+v, err:%v", userKey, err)
		return
	}
	err = logic.RedisPublishChannel(serviceIds, args.ToUserId, bodyByte)
	if err == nil {
		reply.Code = config.SuccessReplyCode
	}
	return
}
