package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/zzjbattlefield/IM_GO/api/rpc"
	"github.com/zzjbattlefield/IM_GO/config"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

type FormRoomInfo struct {
	RoomID int
}

func GetRoomInfo(c *gin.Context) {
	var formInfo = &FormRoomInfo{}
	if err := c.ShouldBindBodyWith(formInfo, binding.JSON); err != nil {
		config.Zap.Errorf("binding form error:%v", err.Error())
	}
	req := &proto.Send{
		Op:     config.OpRoomInfoSend,
		RoomId: formInfo.RoomID,
	}
	code, msg := rpc.RpcLoginObj.GetRoomInfo(req)
	if code == tools.CodeFail {
		config.Zap.Errorf("rpc get room info fail:%v", msg)
		tools.FailWithMessage(c, msg)
		return
	}
	tools.SuccessWithMessage(c, "ok", msg)
	return
}

type FromRoom struct {
	RoomId    int    `json:"roomId" binding:"required" form:"roomId"`
	Msg       string `json:"msg" binding:"required" form:"msg"`
	AuthToken string `json:"authToken" binding:"required" form:"authToken"`
}

func PushRoom(c *gin.Context) {
	var fromRoom FromRoom
	if err := c.ShouldBindBodyWith(&fromRoom, binding.JSON); err != nil {
		tools.FailWithMessage(c, err.Error())
	}
	authToken := fromRoom.AuthToken
	rpc := new(rpc.RpcLogic)
	code, userName, userID := rpc.CheckAuth(&proto.CheckAuthRequest{AuthToken: authToken})
	if code == tools.CodeFail {
		tools.FailWithMessage(c, "错误的authToken数据")
	}
	req := &proto.Send{
		Msg:          fromRoom.Msg,
		FromUserId:   userID,
		FromUserName: userName,
		RoomId:       fromRoom.RoomId,
		Op:           config.OpRoomSend,
	}
	code, msg := rpc.PushRoom(req)
	if code == tools.CodeFail {
		tools.FailWithMessage(c, "pushRoom rpc调用失败")
		return
	}
	tools.SuccessWithMessage(c, "ok", msg)
}

type FromPush struct {
	ToUserId  int    `json:"to_user_id" binding:"required" from:"to_user_id"`
	Msg       string `json:"msg" binding:"required" from:"msg"`
	AuthToken string `json:"authToken" binding:"required" form:"authToken"`
}

func Push(c *gin.Context) {
	fromPush := &FromPush{}
	err := c.ShouldBindBodyWith(fromPush, binding.JSON)
	if err != nil {
		tools.FailWithMessage(c, err.Error())
	}
	userReq := &proto.GetUserInfoRequest{UserId: fromPush.ToUserId}
	code, toUserName := rpc.RpcLoginObj.GetUserInfoByUserId(userReq)
	if code == config.FailReplyCode {
		tools.FailWithMessage(c, "调用GetUserInfoByUserId方法失败")
		return
	}
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: fromPush.AuthToken}
	code, fromUserName, fromUserID := rpc.RpcLoginObj.CheckAuth(checkAuthReq)
	if code == tools.CodeFail {
		tools.FailWithMessage(c, "rpc调用CheckAuth方法失败")
		return
	}
	SendReq := &proto.Send{
		Op:           config.OpSingleSend,
		FromUserId:   fromUserID,
		ToUserId:     fromPush.ToUserId,
		ToUserName:   toUserName,
		FromUserName: fromUserName,
		Msg:          fromPush.Msg,
	}
	code, message := rpc.RpcLoginObj.Push(SendReq)
	if code == config.FailReplyCode {
		tools.FailWithMessage(c, message)
		return
	}
	tools.SuccessWithMessage(c, "ok", nil)
}
