package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzjbattlefield/IM_GO/api/handler"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())
	initUserRouter(r)
	return r
}

// 用户相关路由
func initUserRouter(r *gin.Engine) {
	r.Group("/user")
	r.POST("/login", handler.Login)
	r.POST("/register", handler.Register)
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}
