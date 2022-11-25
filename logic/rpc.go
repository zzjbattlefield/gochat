package logic

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/smallnest/rpcx/server"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/model"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

type LogicRpc struct{}

func (logic *Logic) InitRpcServer() (err error) {
	s := server.NewServer()
	if err = s.RegisterName("LogicRpc", new(LogicRpc), ""); err != nil {
		return err
	}
	err = s.Serve("tcp", "127.0.0.1:6900")
	return err
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

func (rpc *LogicRpc) CheckAuth(ctx context.Context, request *proto.CheckAuthRequest, reply *proto.CheckAuthReponse) (err error) {
	var tokenVal = make(map[string]string)
	reply.Code = tools.CodeFail
	authToken := request.AuthToken
	tokenVal, err = RedisClient.HGetAll(ctx, authToken).Result()
	if err != nil || len(tokenVal) == 0 {
		config.Zap.Errorw("检测authToken失败", "authToken", authToken)
		return
	}
	reply.UserID, _ = strconv.Atoi(tokenVal["userID"])
	reply.UserName = tokenVal["userName"]
	reply.Code = tools.CodeSuccess
	return
}

func CreateAuthToken(ctx context.Context, user *model.UserModel) (authToken string, err error) {
	randStr := tools.GetRandString(32)
	sessionID := tools.CreateSessionId(randStr)
	sessionData := make(map[string]interface{})
	sessionData["userName"] = user.UserName
	sessionData["userID"] = user.ID
	err = RedisClient.HSet(ctx, sessionID, sessionData).Err()
	if err != nil {
		return "", err
	}
	err = RedisClient.Expire(ctx, sessionID, 86400*time.Second).Err()
	if err != nil {
		return "", err
	}
	return sessionID, nil
}
