package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/zzjbattlefield/IM_GO/api/rpc"
	"github.com/zzjbattlefield/IM_GO/proto"
	"github.com/zzjbattlefield/IM_GO/tools"
)

type FormLogin struct {
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
}

type FormRegister struct {
	UserName string `form:"userName" json:"userName" binding:"required"`
	Password string `form:"passWord" json:"passWord" binding:"required"`
}

type FormCheckAuth struct {
	AuthToken string `form:"authToken" json:"authToken" binding:"required"`
}

func Login(c *gin.Context) {
	var formLogin FormLogin
	if err := c.ShouldBindWith(&formLogin, binding.JSON); err != nil {
		tools.FailWithMessage(c, err.Error())
		return
	}
	request := &proto.LoginRequest{
		UserName: formLogin.UserName,
		Password: formLogin.Password,
	}
	code, authToken, msg := rpc.RpcLoginObj.Login(request)
	if code == tools.CodeFail || authToken == "" {
		tools.ResponWithCode(c, tools.CodeFail, msg, nil)
		return
	}
	tools.ResponWithCode(c, tools.CodeSuccess, "login success", authToken)
}

func Register(c *gin.Context) {
	var formRegister FormRegister
	if err := c.ShouldBindWith(&formRegister, binding.JSON); err != nil {
		tools.FailWithMessage(c, err.Error())
		return
	}
	registerRequest := &proto.RegisterRequest{
		UserName: formRegister.UserName,
		Password: formRegister.Password,
	}
	code, authToken, msg := rpc.RpcLoginObj.Register(registerRequest)
	if code == tools.CodeFail || authToken == "" {
		tools.FailWithMessage(c, msg)
		return
	}
	tools.SuccessWithMessage(c, "register success", authToken)
}

func CheckAuth(c *gin.Context) {
	formCheckAuth := &FormCheckAuth{}
	err := c.ShouldBindBodyWith(formCheckAuth, binding.JSON)
	if err != nil {
		tools.FailWithMessage(c, err.Error())
		return
	}

	code, userName, userID := rpc.RpcLoginObj.CheckAuth(&proto.CheckAuthRequest{AuthToken: formCheckAuth.AuthToken})
	if code == tools.CodeFail {
		tools.FailWithMessage(c, "校验失败")
		return
	}
	userData := map[string]interface{}{
		"userName": userName,
		"userID":   userID,
	}
	tools.SuccessWithMessage(c, "获取成功", userData)
}
