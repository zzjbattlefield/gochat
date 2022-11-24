package tools

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess      = 0
	CodeFail         = 1
	CodeUnknowFail   = -1
	CodeSessionError = 40000
)

var MsgCodeMap = map[int]string{
	CodeSuccess:      "success",
	CodeFail:         "fail",
	CodeUnknowFail:   "unknow fail",
	CodeSessionError: "session error",
}

func FailWithMessage(c *gin.Context, msg interface{}) {
	ResponWithCode(c, CodeFail, msg, nil)
}

func SuccessWithMessage(c *gin.Context, msg interface{}, data interface{}) {
	ResponWithCode(c, CodeSuccess, msg, data)
}

func ResponWithCode(c *gin.Context, code int, msg interface{}, data interface{}) {
	if msg == nil {
		if m, ok := MsgCodeMap[code]; ok {
			msg = m
		} else {
			msg = MsgCodeMap[CodeUnknowFail]
		}
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
		"data":    data,
	})
}
